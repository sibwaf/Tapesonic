<script setup lang="ts">
import type { UserRq } from '@/api';
import api from '@/api';
import UserEditor, { type EditableUserProperties, userRsToEditableUserProperties } from '@/components/UserEditor.vue';
import RemoteEditorList from '@/components/RemoteEditorList.vue';
import symbols from '@/symbols';
import { inject, ref, watch, Suspense, computed } from 'vue';
import UserEditorList from '@/components/UserEditorList.vue';
import LastFmSessionEditor from '@/components/LastFmSessionEditor.vue';
import ListenBrainzSessionEditor from '@/components/ListenBrainzSessionEditor.vue';

const { currentUser, updateCurrentUser, updateCurrentUserApiKey } = inject(symbols.currentUser)!;
const editableCurrentUser = ref<EditableUserProperties | null>(null);

watch(currentUser, (newValue) => {
    if (newValue == null) {
        editableCurrentUser.value = null;
    } else {
        editableCurrentUser.value = userRsToEditableUserProperties(newValue);
    }
}, { immediate: true });

const isBusy = ref(false);

async function onUpdateCurrentUser(user: EditableUserProperties) {
    try {
        isBusy.value = true;

        const rq: UserRq = {
            Name: user.editedValue.name,
            Role: null,
            Password: user.editedValue.password,
        };
        if (!user.isPasswordEdited) {
            rq.Password = null;
        }

        const rs = await api.patchUser(currentUser.value!.Id, rq)
        updateCurrentUser(rs);
    } catch (e) {
        console.error("Failed to patch user", e);
    } finally {
        isBusy.value = false;
    }
}

async function onRegenerateCurrentUserApiKey(userId: string) {
    try {
        isBusy.value = true;

        const rs = await api.postUserApiKeys(userId);
        updateCurrentUserApiKey(rs.ApiKey);
    } catch (e) {
        console.error("Failed to regenerate api key", e);
    } finally {
        isBusy.value = false;
    }
}

watch(computed(() => currentUser.value?.ApiKey), (apiKey) => {
    const editableCurrentUserValue = editableCurrentUser.value;
    if (apiKey != null && editableCurrentUserValue != null) {
        editableCurrentUserValue.editedValue.apiKey = apiKey;
    }
});
</script>

<template>
    <Suspense>
        <div v-if="currentUser != null">
            <template v-if="editableCurrentUser">
                <h2>Account</h2>

                <UserEditor :user="editableCurrentUser" :disabled="isBusy" :allow-role-editing="false"
                    :require-username="false" :require-password="false" @save="onUpdateCurrentUser(editableCurrentUser)"
                    @regenerate-api-key="onRegenerateCurrentUserApiKey(currentUser.Id)" />
            </template>

            <h2>ListenBrainz</h2>

            <ListenBrainzSessionEditor />

            <h2>last.fm</h2>

            <LastFmSessionEditor />

            <template v-if="currentUser?.Role == 'ADMIN'">
                <h2>Users</h2>

                <UserEditorList :current-user="currentUser" />
            </template>

            <h2>Remotes</h2>

            <RemoteEditorList :allow-config-change="currentUser?.Role == 'ADMIN'" />
        </div>
        <div v-else>
            Not authenticated?
        </div>

        <template #fallback>
            Loading...
        </template>
    </Suspense>
</template>
