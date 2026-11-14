package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"ecommerce-backend/internal/models"

	"github.com/shopspring/decimal"
)

type OrderRepository struct {
	DB *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{
		DB: db,
	}
}

func (r *OrderRepository) CreateOrder(
	userID int64,
	addressID int64,
) (models.Order, error) {

	var order models.Order

	tx, err := r.DB.Begin()
	if err != nil {
		return order, err
	}
	defer tx.Rollback()

	// 1. Lock and retrieve the customer's address.
	var address models.Address

	err = tx.QueryRow(`
		SELECT id, recipient_name, phone, city, area,
		       street, building_number, apartment_number,
		       additional_instructions
		FROM user_addresses
		WHERE id = $1 AND user_id = $2
		FOR UPDATE
	`, addressID, userID).Scan(
		&address.ID,
		&address.RecipientName,
		&address.Phone,
		&address.City,
		&address.Area,
		&address.Street,
		&address.BuildingNumber,
		&address.ApartmentNumber,
		&address.AdditionalInstructions,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return order, errors.New("delivery address not found")
		}
		return order, err
	}

	// 2. Lock and retrieve the customer's cart.
	var cartID int64

	err = tx.QueryRow(`
		SELECT id
		FROM carts
		WHERE user_id = $1
		FOR UPDATE
	`, userID).Scan(&cartID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return order, errors.New("cart not found")
		}
		return order, err
	}

	// 3. Retrieve cart items and lock their product rows.
	rows, err := tx.Query(`
	SELECT
		ci.product_id,
		p.name,
		p.name_ar,
		ci.quantity,
		p.price,
		p.discount_value,
		ci.selected_color,
		pc.name_ar,
		p.stock,
		p.is_active
	FROM cart_items ci
	JOIN products p ON p.id = ci.product_id
	LEFT JOIN product_colors pc
		ON pc.product_id = p.id
		AND pc.name = ci.selected_color
	WHERE ci.cart_id = $1
	ORDER BY ci.product_id
	FOR UPDATE OF p
`, cartID)

	if err != nil {
		return order, err
	}

	var items []models.OrderItem
	total := decimal.Zero

	for rows.Next() {
		var item models.OrderItem
		var price decimal.Decimal
		var discount decimal.Decimal
		var stock int
		var isActive bool

		err := rows.Scan(
			&item.ProductID,
			&item.ProductName,
			&item.ProductNameAr,
			&item.Quantity,
			&price,
			&discount,
			&item.SelectedColor,
			&item.SelectedColorAr,
			&stock,
			&isActive,
		)
		if err != nil {
			return order, err
		}

		if !isActive {
			return order, fmt.Errorf(
				"product %s is no longer available",
				item.ProductName,
			)
		}
		if item.SelectedColor != nil && item.SelectedColorAr == nil {
			return order, fmt.Errorf(
				"selected color is unavailable for %s",
				item.ProductName,
			)
		}

		if item.Quantity > stock {
			return order, fmt.Errorf(
				"insufficient stock for %s",
				item.ProductName,
			)
		}

		// Apply percentage discount.
		discountMultiplier := decimal.NewFromInt(100).Sub(discount)

		item.UnitPrice = price.
			Mul(discountMultiplier).
			Div(decimal.NewFromInt(100)).
			Round(2)

		item.Subtotal = item.UnitPrice.Mul(
			decimal.NewFromInt(int64(item.Quantity)),
		).Round(2)

		total = total.Add(item.Subtotal)

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return order, err
	}

	if err := rows.Close(); err != nil {
		return order, err
	}

	if len(items) == 0 {
		return order, errors.New("cart is empty or contains no active products")
	}

	// 4. Create the order using a snapshot of the address.
	err = tx.QueryRow(`
		INSERT INTO orders (
			user_id, address_id, recipient_name, phone,
			city, area, street, building_number,
			apartment_number, additional_instructions,
			total_amount, payment_method, payment_status
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, 'cash_on_delivery', 'pending'
		)
		RETURNING id, status, payment_method, payment_status,
		          created_at, updated_at
	`,
		userID,
		addressID,
		address.RecipientName,
		address.Phone,
		address.City,
		address.Area,
		address.Street,
		address.BuildingNumber,
		address.ApartmentNumber,
		address.AdditionalInstructions,
		total,
	).Scan(
		&order.ID,
		&order.Status,
		&order.PaymentMethod,
		&order.PaymentStatus,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err != nil {
		return order, err
	}

	// 5. Insert order items and reduce product stock.
	for i := range items {
		item := &items[i]
		item.OrderID = order.ID

		err = tx.QueryRow(`
			INSERT INTO order_items (
				order_id, product_id, product_name,
				product_name_ar, quantity, unit_price,
				selected_color, selected_color_ar, subtotal
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id
		`,
			item.OrderID,
			item.ProductID,
			item.ProductName,
			item.ProductNameAr,
			item.Quantity,
			item.UnitPrice,
			item.SelectedColor,
			item.SelectedColorAr,
			item.Subtotal,
		).Scan(&item.ID)

		if err != nil {
			return order, err
		}

		result, err := tx.Exec(`
	UPDATE products
	SET stock = stock - $1,
	    updated_at = NOW()
	WHERE id = $2
	  AND stock >= $1
	  AND is_active = TRUE
`, item.Quantity, item.ProductID)

		if err != nil {
			return order, err
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return order, err
		}

		if affected != 1 {
			return order, fmt.Errorf(
				"unable to update stock for %s",
				item.ProductName,
			)
		}
	}

	// 6. Clear the cart.
	_, err = tx.Exec(`
		DELETE FROM cart_items
		WHERE cart_id = $1
	`, cartID)

	if err != nil {
		return order, err
	}

	// 7. Commit everything together.
	if err := tx.Commit(); err != nil {
		return order, err
	}

	order.UserID = userID
	order.AddressID = &addressID
	order.RecipientName = address.RecipientName
	order.Phone = address.Phone
	order.City = address.City
	order.Area = address.Area
	order.Street = address.Street
	order.BuildingNumber = address.BuildingNumber
	order.ApartmentNumber = address.ApartmentNumber
	order.AdditionalInstructions = address.AdditionalInstructions
	order.TotalAmount = total
	order.Items = items

	return order, nil
}
func (r *OrderRepository) GetOrdersByUserID(
	userID int64,
) ([]models.Order, error) {

	rows, err := r.DB.Query(`
		SELECT
			id, user_id, address_id,
			recipient_name, phone, city, area, street,
			building_number, apartment_number,
			additional_instructions, total_amount,
			status, payment_method, payment_status,
			created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)

	if err != nil {
		return nil, err
	}

	var orders []models.Order

	for rows.Next() {
		var order models.Order

		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.AddressID,
			&order.RecipientName,
			&order.Phone,
			&order.City,
			&order.Area,
			&order.Street,
			&order.BuildingNumber,
			&order.ApartmentNumber,
			&order.AdditionalInstructions,
			&order.TotalAmount,
			&order.Status,
			&order.PaymentMethod,
			&order.PaymentStatus,
			&order.CreatedAt,
			&order.UpdatedAt,
		)

		if err != nil {
			rows.Close()
			return nil, err
		}

		order.Items = []models.OrderItem{}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	// Retrieve the items belonging to each order.
	for i := range orders {
		itemRows, err := r.DB.Query(`
			SELECT
				id, order_id, product_id,
				product_name, product_name_ar,
				quantity, unit_price,
				selected_color, selected_color_ar,
				subtotal
			FROM order_items
			WHERE order_id = $1
			ORDER BY id
		`, orders[i].ID)

		if err != nil {
			return nil, err
		}

		for itemRows.Next() {
			var item models.OrderItem

			err := itemRows.Scan(
				&item.ID,
				&item.OrderID,
				&item.ProductID,
				&item.ProductName,
				&item.ProductNameAr,
				&item.Quantity,
				&item.UnitPrice,
				&item.SelectedColor,
				&item.SelectedColorAr,
				&item.Subtotal,
			)

			if err != nil {
				itemRows.Close()
				return nil, err
			}

			orders[i].Items = append(orders[i].Items, item)
		}

		if err := itemRows.Err(); err != nil {
			itemRows.Close()
			return nil, err
		}

		if err := itemRows.Close(); err != nil {
			return nil, err
		}
	}

	return orders, nil
}

func (r *OrderRepository) GetOrderByID(
	userID int64,
	orderID int64,
) (models.Order, error) {

	var order models.Order

	err := r.DB.QueryRow(`
		SELECT
			id, user_id, address_id,
			recipient_name, phone, city, area, street,
			building_number, apartment_number,
			additional_instructions, total_amount,
			status, payment_method, payment_status,
			created_at, updated_at
		FROM orders
		WHERE id = $1 AND user_id = $2
	`, orderID, userID).Scan(
		&order.ID,
		&order.UserID,
		&order.AddressID,
		&order.RecipientName,
		&order.Phone,
		&order.City,
		&order.Area,
		&order.Street,
		&order.BuildingNumber,
		&order.ApartmentNumber,
		&order.AdditionalInstructions,
		&order.TotalAmount,
		&order.Status,
		&order.PaymentMethod,
		&order.PaymentStatus,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err != nil {
		return order, err
	}

	order.Items = []models.OrderItem{}

	rows, err := r.DB.Query(`
		SELECT
			id, order_id, product_id,
			product_name, product_name_ar,
			quantity, unit_price,
			selected_color, selected_color_ar,
			subtotal
		FROM order_items
		WHERE order_id = $1
		ORDER BY id
	`, order.ID)

	if err != nil {
		return order, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.OrderItem

		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.ProductName,
			&item.ProductNameAr,
			&item.Quantity,
			&item.UnitPrice,
			&item.SelectedColor,
			&item.SelectedColorAr,
			&item.Subtotal,
		)

		if err != nil {
			return order, err
		}

		order.Items = append(order.Items, item)
	}

	if err := rows.Err(); err != nil {
		return order, err
	}

	return order, nil
}
func (r *OrderRepository) CancelOrder(
	userID int64,
	orderID int64,
) error {

	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Lock the order and check ownership.
	var status string

	err = tx.QueryRow(`
		SELECT status
		FROM orders
		WHERE id = $1 AND user_id = $2
		FOR UPDATE
	`, orderID, userID).Scan(&status)

	if err != nil {
		return err
	}

	if status != "pending" {
		return errors.New("only pending orders can be cancelled")
	}

	// 2. Retrieve the order items.
	rows, err := tx.Query(`
		SELECT product_id, quantity
		FROM order_items
		WHERE order_id = $1
	`, orderID)

	if err != nil {
		return err
	}

	type cancelledItem struct {
		ProductID sql.NullInt64
		Quantity  int
	}

	var items []cancelledItem

	for rows.Next() {
		var item cancelledItem

		if err := rows.Scan(&item.ProductID, &item.Quantity); err != nil {
			rows.Close()
			return err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}

	if err := rows.Close(); err != nil {
		return err
	}

	// 3. Restore product stock.
	for _, item := range items {
		if !item.ProductID.Valid {
			continue
		}

		_, err = tx.Exec(`
			UPDATE products
			SET stock = stock + $1,
			    updated_at = NOW()
			WHERE id = $2
		`, item.Quantity, item.ProductID.Int64)

		if err != nil {
			return err
		}
	}

	// 4. Update the order status.
	_, err = tx.Exec(`
		UPDATE orders
		SET status = 'cancelled',
		    updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`, orderID, userID)

	if err != nil {
		return err
	}

	return tx.Commit()
}
func (r *OrderRepository) UpdateOrderStatus(orderID int64, newStatus string) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var currentStatus string

	err = tx.QueryRow(`
        SELECT status
        FROM orders
        WHERE id = $1
        FOR UPDATE
    `, orderID).Scan(&currentStatus)

	if err != nil {
		return err
	}

	validTransition := false

	switch currentStatus {
	case "pending":
		validTransition = newStatus == "confirmed"
	case "confirmed":
		validTransition = newStatus == "processing"
	case "processing":
		validTransition = newStatus == "shipped"
	case "shipped":
		validTransition = newStatus == "delivered"
	}

	if !validTransition {
		return fmt.Errorf(
			"cannot change order status from %s to %s",
			currentStatus,
			newStatus,
		)
	}

	_, err = tx.Exec(`
        UPDATE orders
        SET status = $1,
            updated_at = NOW()
        WHERE id = $2
    `, newStatus, orderID)

	if err != nil {
		return err
	}

	return tx.Commit()
}
