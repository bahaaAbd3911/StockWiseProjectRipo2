# 11. Pay later and customer debts

## 11.1 What it is

Trusted customers can buy now and pay later. The store keeps a ledger per
customer, like a debt notebook.

1. An admin gives a customer a **credit limit**, e.g. 500.00. The default is
   0, which means pay later is not offered.
2. At checkout the customer chooses **Pay Later**. If
   `balance + order total ≤ credit limit`, the order is placed with
   `payment_status: deferred`, and the total is added to the debt.
3. The admin records money as the customer pays it back.
4. A cash order can also become debt: the admin marks it delivered with
   `payment_collected: false`.

## 11.2 The ledger

Every change is **one row** in `customer_debt_transactions`. Rows are never
edited or deleted; a mistake is corrected with a new row.
**Balance = sum of all amounts.**

| type | sign | created when |
|---|---|---|
| `charge` | + | pay-later checkout (automatic); cash order delivered unpaid (admin) |
| `payment` | − | admin records money received |
| `reversal` | − | a charged order is cancelled (automatic) |
| `adjustment` | + or − | admin correction; **a note is required** |

Each row is shown with `balance_after`, the running balance, so the list
reads like a bank statement.

**Why append-only, not a "balance" column?** A balance column that gets
overwritten loses the story: nobody can see why it is 40.00. A ledger keeps
every charge and payment, and corrections are visible instead of silent.

The rules:

- A payment cannot be larger than the balance; the store does not keep
  credit in the customer's favour.
- An adjustment cannot make the balance negative.
- Lowering a credit limit below the current balance is allowed; it only
  blocks new pay-later orders.
- Cancelling a pay-later order removes its remaining charge. Payments
  already made are kept, because that money was really received.
- **Everything that changes a balance first locks the customer's row**, so
  two checkouts at the same moment cannot both use the same credit.
- `CHECK` constraints enforce the sign per type and the note on
  adjustments.

## 11.3 Payment status

| Order | `payment_method` | `payment_status` |
|---|---|---|
| cash, placed | `cash_on_delivery` | `pending` |
| cash, delivered and paid | `cash_on_delivery` | `paid` |
| cash, delivered, not paid | `cash_on_delivery` | `deferred` (now in the debt) |
| pay later | `pay_later` | `deferred` |

Debt is tracked **per customer**, not per order. A payment lowers the
customer's balance, not one order's status.

## 11.4 Emails

- **Payment received**: the amount and the remaining balance.
- **Balance corrected**: the amount, the reason and the new balance.
- **Reminders**: while a balance is open, every `DEBT_REMINDER_INTERVAL`
  (default weekly). The reminder and the "reminded at" date are written in
  one transaction, so nobody is reminded twice for the same period.
  Paid-up customers get nothing.

## 11.5 The screens

- **Customers (Profile → My Balance):** what they owe, their credit limit,
  what is still available to pay later, and every transaction. If pay later
  is off, the page says so.
- **Checkout:** the "Pay Later" option appears only when allowed, with the
  available credit in its description.
- **Admins (Customer Debts):**
  - totals: total owed, customers owing, number of accounts;
  - a list sorted by biggest balance, with an "only customers who owe money"
    filter;
  - "Allow a customer": pick a customer and give them a limit;
  - one customer's statement: change the limit, record a payment, record a
    correction (note required), and see every transaction with the running
    balance.
