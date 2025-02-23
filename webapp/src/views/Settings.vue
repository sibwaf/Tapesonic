<script setup lang="ts">
import type { LastFmAuthLinkRs, LastFmSessionRs } from '@/api';
import api from '@/api';
import { onMounted, ref } from 'vue';

const isBusy = ref(false);
const isLoaded = ref(false);

const lastFmSession = ref<LastFmSessionRs | null>();
const pendingLastFmAuth = ref<LastFmAuthLinkRs | null>();

async function authorizeLastFm() {
    try {
        isBusy.value = true;
        pendingLastFmAuth.value = null;
        pendingLastFmAuth.value = await api.createLastFmAuthLink();
    } catch (e) {
        console.error("Failed to get an authentication link for last.fm", e);
    } finally {
        isBusy.value = false;
    }
}

async function createLastFmSession() {
    try {
        isBusy.value = true;
        lastFmSession.value = await api.createLastFmSession(pendingLastFmAuth.value!.Token);
    } catch (e) {
        console.error("Failed to create a last.fm session", e);
    } finally {
        isBusy.value = false;
    }
}

onMounted(async () => {
    try {
        isBusy.value = true;
        lastFmSession.value = await api.getCurrentLastFmSession();
    } catch (e) {
        console.error("Failed to fetch current last.fm session", e);
    } finally {
        isBusy.value = false;
        isLoaded.value = true;
    }
});
</script>

<template>
    <div v-if="isLoaded">
        <h2>last.fm</h2>

        <div v-if="lastFmSession">Current user: {{ lastFmSession.Username }}</div>
        <div v-else>Not authorized</div>

        <button @click="authorizeLastFm" :disabled="isBusy">Connect to last.fm</button>
        <ol v-if="pendingLastFmAuth">
            <li><a :href="pendingLastFmAuth.Url" target="_blank">Authorize Tapesonic to access your last.fm account</a></li>
            <li><button @click="createLastFmSession" :disabled="isBusy">Create session</button></li>
        </ol>
    </div>
    <div v-else>
        Loading...
    </div>
</template>
