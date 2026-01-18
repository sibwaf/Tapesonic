<script lang="ts">
export interface RemoteConfig {
    name: string;
    type: string;
    url: string;
    isScrobbleReplicationEnabled: boolean;
    isExternalScrobblingEnabled: boolean;
}

export interface RemoteCredentials {
    username: string;
    password: string;
}

export interface RemoteInfo {
    status: string;
    description: string;
}

export class EditableRemoteConfig implements Editable<RemoteConfig> {

    private state: RemoteConfig;

    private static makeModified(config: RemoteConfig): RemoteConfig {
        return JSON.parse(JSON.stringify(config));
    }

    public constructor(private readonly original: RemoteConfig) {
        this.state = EditableRemoteConfig.makeModified(original);
    }

    public get editedValue(): RemoteConfig {
        return this.state;
    }

    public get isEdited(): boolean {
        return JSON.stringify(this.original) != JSON.stringify(this.editedValue);
    }

    public reset(): void {
        this.state = EditableRemoteConfig.makeModified(this.original);
    }
}
</script>

<script setup lang="ts">
import type { Editable } from '@/model/Editable';
import { computed, ref, watch } from 'vue';

const props = defineProps<{
    id: string | null,
    config: Editable<RemoteConfig>,
    credentials: RemoteCredentials,
    info: RemoteInfo,
    allowConfigChange: boolean,
    disabled: boolean,
}>();
const emit = defineEmits<{
    save: [],
    delete: [],
    authenticate: [credentials: any],
    deauthenticate: [],
}>();

const isConfigValid = computed(() => {
    return props.config.editedValue.name.trim().length > 0
        && props.config.editedValue.url.trim().length > 0;
})

const username = ref("");
const password = ref("");

const areCredentialsValid = computed(() => {
    return username.value.trim().length > 0
        && password.value.length > 0;
});

watch(computed(() => props.credentials.username), (newValue) => {
    username.value = newValue;
    password.value = "";
}, { immediate: true });

function onSave() {
    emit("save");
}

function onDelete() {
    emit("delete");
}

function onAuthenticate() {
    const credentials = {
        Username: username.value.trim(),
        Password: password.value,
    };

    emit("authenticate", credentials);
}

function onDeauthenticate() {
    emit("deauthenticate");
}

const statusText = computed(() => {
    switch (props.info.status) {
        case "BAD_RESPONSE": return "Bad response from remote";
        case "UNREACHABLE": return "Remote is not reachable";
        case "NOT_AUTHENTICATED": return "Not authenticated";
        case "UNKNOWN": return "Failed to get remote status";
    }

    return props.info.status;
});
</script>

<template>
    <div>
        <div>
            <input type="text" placeholder="Name" v-model="config.editedValue.name"
                :disabled="disabled || !allowConfigChange">
        </div>
        <div>
            <input type="text" placeholder="URL" v-model="config.editedValue.url"
                :disabled="disabled || !allowConfigChange">
        </div>
        <div>
            <input type="checkbox" v-model="config.editedValue.isScrobbleReplicationEnabled"
                :disabled="disabled || !allowConfigChange">
            Tapesonic should report scrobbles to this remote
        </div>
        <div>
            <input type="checkbox" v-model="config.editedValue.isExternalScrobblingEnabled"
                :disabled="disabled || !allowConfigChange">
            Tapesonic should scrobble tracks hosted on this remote to external services (ex. ListenBrainz)
        </div>

        <div v-if="allowConfigChange">
            <button :disabled="disabled || !config.isEdited || !isConfigValid" @click="onSave">Save</button>
            <button v-if="id != null" :disabled="disabled" @click="onDelete">Delete</button>
        </div>

        <div v-if="id != null">
            <br>

            <div>
                <input type="text" placeholder="Username" v-model="username">
            </div>
            <div>
                <input type="password" placeholder="Password" v-model="password">
            </div>
            <div>
                <button :disabled="disabled || !areCredentialsValid" @click="onAuthenticate">Connect</button>
                <button :disabled="disabled" @click="onDeauthenticate">Disconnect</button>
            </div>
        </div>

        <div v-if="info.description || statusText">
            <br>

            <span v-if="info.description">{{ info.description }}</span>
            <template v-if="info.description && statusText">, </template>
            <span v-if="statusText">{{ statusText }}</span>
        </div>
    </div>
</template>
