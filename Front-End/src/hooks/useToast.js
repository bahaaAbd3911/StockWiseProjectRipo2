import { useContext, useMemo } from "react";
import { AuthContext } from "../context/AuthContext";

// Thin wrapper around the app-wide <Message /> toast. The returned object is
// stable (same identity between renders), so it is safe in effect dependencies.
export function useToast() {
  const { setMessage } = useContext(AuthContext);

  return useMemo(() => {
    const show = (msgTxt, arbMsgTxt, isError = false) =>
      setMessage({
        isShow: true,
        msgTxt,
        arbMsgTxt,
        msgBGColor: isError ? "#EF4444" : "#10BBB1",
        time: isError ? 5000 : 3000,
        reminder: isError,
      });
    return {
      success: (en, ar) => show(en, ar),
      error: (en, ar) => show(en, ar, true),
    };
  }, [setMessage]);
}
