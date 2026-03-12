<script setup lang="ts">
import { provide, ref } from "vue";
import { RouterView, useRouter } from "vue-router"
import usersApi, { type UserRs } from "@/api/users";
import symbols from "@/symbols";

const router = useRouter();

const currentUser = ref<UserRs | null>(null);

function updateCurrentUser(user: UserRs | null) {
    currentUser.value = user;
}

function updateCurrentUserApiKey(apiKey: string) {
    const currentUserValue = currentUser.value;
    if (currentUserValue == null) {
        return;
    }

    currentUserValue.ApiKey = apiKey;
}

provide(symbols.currentUser, { currentUser, updateCurrentUser, updateCurrentUserApiKey });

router.beforeEach(async (to) => {
    if (to.name == "setup") {
        return true;
    }

    if (currentUser.value == null) {
        try {
            updateCurrentUser(await usersApi.getCurrentUser());
        } catch (e) {
            console.log("Failed to retrieve current user", e)
        }
    }

    return true;
});
</script>

<template>
    <span class="header-link">
        <RouterLink to="/">Home</RouterLink>
    </span>
    <span class="header-link">
        <RouterLink to="/artists">Artists</RouterLink>
    </span>
    <span class="header-link">
        <RouterLink to="/sources">Sources</RouterLink>
    </span>
    <span class="header-link">
        <RouterLink to="/settings">Settings</RouterLink>
    </span>
    <span class="header-link">
        <RouterLink to="/tapes/new">New tape</RouterLink>
    </span>

    <hr>

    <RouterView />
</template>

<style>
@import "primeicons/primeicons.css";

.header-link {
    padding: 0.25em;
}
</style>
