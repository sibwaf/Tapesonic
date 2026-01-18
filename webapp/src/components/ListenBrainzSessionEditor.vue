<script setup lang="ts">
import api, { type ListenBrainzSessionRs, type ListenBrainzSessionSettingsRq } from '@/api';
import { computed, ref, watch } from 'vue';

const isBusy = ref(false);

const token = ref("");

const session = ref<ListenBrainzSessionRs | null>();
const sessionSettings = ref<ListenBrainzSessionSettingsRq | null>();

const sessionSettingsChanged = computed(() => {
    return session?.value?.IsScrobblingEnabled != sessionSettings?.value?.IsScrobblingEnabled;
});

watch(session, newValue => {
    if (newValue == null) {
        sessionSettings.value = null;
    } else {
        sessionSettings.value = {
            IsScrobblingEnabled: newValue.IsScrobblingEnabled,
        };
    }
});

async function createSession() {
    try {
        isBusy.value = true;
        session.value = await api.createListenBrainzSession(token.value);
        token.value = "";
    } catch (e) {
        console.error("Failed to create a ListenBrainz session", e);
    } finally {
        isBusy.value = false;
    }
}

async function updateSessionSettings() {
    const settingsValue = sessionSettings.value;
    if (settingsValue == null) {
        return;
    }

    try {
        isBusy.value = true;
        session.value = await api.updateListenBrainzSessionSettings(settingsValue);
    } catch (e) {
        console.error("Failed to update ListenBrainz session settings", e);
    } finally {
        isBusy.value = false;
    }
}

async function deleteSession() {
    try {
        isBusy.value = true;
        await api.deleteListenBrainzSession();
        session.value = null;
    } catch (e) {
        console.error("Failed to delete the ListenBrainz session", e)
    } finally {
        isBusy.value = false;
    }
}

try {
    isBusy.value = true;
    session.value = await api.getCurrentListenBrainzSession();
} catch (e) {
    console.error("Failed to fetch current ListenBrainz session", e);
} finally {
    isBusy.value = false;
}
</script>

<template>
    <div>
        <div v-if="session != null && sessionSettings != null">
            <div>
                Current user: {{ session.Username }}
                <button @click="deleteSession" :disabled="isBusy">Log out</button>
            </div>

            <br>

            <div>
                <div>
                    <input type="checkbox" v-model="sessionSettings.IsScrobblingEnabled">
                    Enable scrobbling
                </div>
                <button @click="updateSessionSettings" :disabled="isBusy || !sessionSettingsChanged">Save</button>
            </div>

            <br>
        </div>
        <div v-else>Not authenticated</div>

        <input type="text" placeholder="Token" v-model="token">
        <button @click="createSession" :disabled="!token || isBusy">
            Connect to ListenBrainz
        </button>
    </div>
</template>
