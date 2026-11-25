import { updateOrderStatus } from "../api/admin";
import { useApiAction } from "./useApiAction";

// Changes an order's status on the server, then runs `afterChange` (usually a reload).
export function useOrderStatusUpdate(afterChange) {
  const { run, busy } = useApiAction();

  const changeStatus = (orderId, status) =>
    run(
      async () => {
        await updateOrderStatus(orderId, status);
        await afterChange?.();
      },
      ["Order status updated", "تم تحديث حالة الطلب"],
    );

  return { changeStatus, busy };
}
