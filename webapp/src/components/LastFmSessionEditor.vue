<script setup lang="ts">
import lastfmApi, { type LastFmAuthLinkRs, type LastFmSessionRs, type LastFmSessionSettingsRq } from '@/api/lastfm';
import { computed, ref, watch } from 'vue';

const isBusy = ref(false);

const session = ref<LastFmSessionRs | null>();
const sessionSettings = ref<LastFmSessionSettingsRq | null>();

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

const pendingAuthLink = ref<LastFmAuthLinkRs | null>();

async function startAuthFlow() {
    try {
        isBusy.value = true;
        pendingAuthLink.value = null;
        pendingAuthLink.value = await lastfmApi.createLastFmAuthLink();
    } catch (e) {
        console.error("Failed to get an authentication link for last.fm", e);
    } finally {
        isBusy.value = false;
    }
}

async function createSession() {
    try {
        isBusy.value = true;
        session.value = await lastfmApi.createLastFmSession(pendingAuthLink.value!.Token);
        pendingAuthLink.value = null;
    } catch (e) {
        console.error("Failed to create a last.fm session", e);
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
        session.value = await lastfmApi.updateLastFmSessionSettings(settingsValue);
    } catch (e) {
        console.error("Failed to update last.fm session settings", e);
    } finally {
        isBusy.value = false;
    }
}

async function deleteSession() {
    try {
        isBusy.value = true;
        await lastfmApi.deleteLastFmSession();
        session.value = null;
    } catch (e) {
        console.error("Failed to delete the last.fm session", e)
    } finally {
        isBusy.value = false;
    }
}

try {
    isBusy.value = true;
    session.value = await lastfmApi.getCurrentLastFmSession();
} catch (e) {
    console.error("Failed to fetch current last.fm session", e);
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

        <button @click="startAuthFlow" :disabled="isBusy">Connect to last.fm</button>
        <ol v-if="pendingAuthLink">
            <li>
                <a :href="pendingAuthLink.Url" target="_blank">Authorize Tapesonic to access your last.fm account</a>
            </li>
            <li><button @click="createSession" :disabled="isBusy">Create session</button></li>
        </ol>
    </div>
</template>
