import { useState } from "react";
import { useToast } from "./useToast";

// Wraps an async API call: tracks `busy`, toasts success, and toasts the server's error.
// The wrapped function resolves to true on success and false on failure.
export function useApiAction() {
  const toast = useToast();
  const [busy, setBusy] = useState(false);

  const run = async (action, success) => {
    setBusy(true);
    try {
      await action();
      if (success) toast.success(success[0], success[1]);
      return true;
    } catch (err) {
      toast.error(err.message, err.message);
      return false;
    } finally {
      setBusy(false);
    }
  };

  return { run, busy };
}
