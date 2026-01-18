import type { InjectionKey, Ref } from "vue";
import type { UserRs } from "./api";

export default {
    currentUser: Symbol() as InjectionKey<{
        currentUser: Ref<UserRs | null>,
        updateCurrentUser: (user: UserRs | null) => void,
        updateCurrentUserApiKey: (apiKey: string) => void,
    }>,
};
