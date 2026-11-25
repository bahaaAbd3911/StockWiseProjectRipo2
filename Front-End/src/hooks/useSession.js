import { useCallback, useEffect, useState } from "react";

import * as authApi from "../api/auth";
import { setUnauthorizedHandler, tokenStore } from "../api/client";
import { toUser } from "../api/mappers";

// Signed-in user + token handling. `ready` is false until a saved session has been checked.
export function useSession() {
  const [user, setUser] = useState(null);
  const [ready, setReady] = useState(() => !tokenStore.get());

  const signOut = useCallback(() => {
    tokenStore.clear();
    setUser(null);
  }, []);

  const refresh = useCallback(async () => {
    if (!tokenStore.get()) return;
    setUser(toUser(await authApi.getProfile()));
  }, []);

  useEffect(() => {
    setUnauthorizedHandler(signOut);
    if (!tokenStore.get()) return;

    authApi
      .getProfile()
      .then((profile) => setUser(toUser(profile)))
      .catch(signOut)
      .finally(() => setReady(true));
  }, [signOut]);

  const startSession = ({ token, user: apiUser }) => {
    tokenStore.set(token);
    const mapped = toUser(apiUser);
    setUser(mapped);
    return mapped;
  };

  return {
    user,
    ready,
    isSign: Boolean(user),
    isAdmin: user?.role === "admin",
    signIn: async (email, password) => startSession(await authApi.login(email, password)),
    signUp: async (data) => startSession(await authApi.register(data)),
    signOut,
    updateProfile: async (data) => {
      const updated = toUser(await authApi.updateProfile(data));
      setUser(updated);
      return updated;
    },
    // Reloads the profile, e.g. after the email was confirmed in another tab.
    refreshProfile: refresh,
  };
}
