<script setup lang="ts">
import remotesApi, { type RemoteFullRs, type RemoteListRs, type RemoteRq } from '@/api/remotes';
import { ref } from 'vue';
import RemoteEditor, { EditableRemoteConfig, type RemoteConfig, type RemoteCredentials, type RemoteInfo } from '@/components/RemoteEditor.vue';
import type { Editable } from '@/model/Editable';

const props = defineProps<{
    allowConfigChange: boolean,
}>();

const isBusy = ref(false);

interface Entry {
    id: string;
    config: Editable<RemoteConfig>;
    credentials: RemoteCredentials;
    info: RemoteInfo;
}

function createBlankEntry(): Entry {
    return {
        id: "00000000-0000-0000-0000-000000000000",
        config: new EditableRemoteConfig({
            type: "subsonic",
            name: "",
            url: "",
            isScrobbleReplicationEnabled: true,
            isExternalScrobblingEnabled: true,
        }),
        credentials: { username: "", password: "" },
        info: { description: "", status: "" },
    };
}

function toRemoteConfig(remoteRs: RemoteListRs): RemoteConfig {
    return {
        name: remoteRs.Name,
        type: remoteRs.Type,
        url: remoteRs.Url,
        isScrobbleReplicationEnabled: remoteRs.IsScrobbleReplicationEnabled,
        isExternalScrobblingEnabled: remoteRs.IsExternalScrobblingEnabled,
    };
}

function toRemoteRq(config: RemoteConfig): RemoteRq {
    return {
        Name: config.name,
        Type: config.type,
        Url: config.url,
        IsScrobbleReplicationEnabled: config.isScrobbleReplicationEnabled,
        IsExternalScrobblingEnabled: config.isExternalScrobblingEnabled,
    };
}

const blankRemote = ref<Entry>(createBlankEntry());
const remotes = ref<Entry[]>([]);

async function onCreate(remote: Editable<RemoteConfig>) {
    try {
        isBusy.value = true;

        const createdRemote = await remotesApi.createRemote(toRemoteRq(remote.editedValue));

        replaceRemoteState(createdRemote);
        blankRemote.value = createBlankEntry();
    } catch (e) {
        console.error("Failed to create remote", e);
    } finally {
        isBusy.value = false;
    }
}

async function onUpdate(id: string, remote: Editable<RemoteConfig>) {
    try {
        isBusy.value = true;

        const updatedRemote = await remotesApi.updateRemote(id, toRemoteRq(remote.editedValue));
        replaceRemoteState(updatedRemote);
    } catch (e) {
        console.error("Failed to update remote", e);
    } finally {
        isBusy.value = false;
    }
}

async function onDelete(id: string) {
    try {
        isBusy.value = true;

        await remotesApi.deleteRemote(id);

        const index = remotes.value.findIndex(it => it.id == id);
        if (index >= 0) {
            remotes.value.splice(index, 1);
        }
    } catch (e) {
        console.error("Failed to delete remote", e);
    } finally {
        isBusy.value = false;
    }
}

async function onAuthenticate(id: string, credentials: any) {
    try {
        isBusy.value = true;
        const result = await remotesApi.authenticateRemote(id, credentials);
        replaceRemoteState(result);
    } catch (e) {
        console.error("Failed to authenticate remote", e);
    } finally {
        isBusy.value = false;
    }
}

async function onDeauthenticate(id: string) {
    try {
        isBusy.value = true;
        const result = await remotesApi.deauthenticateRemote(id);
        replaceRemoteState(result);
    } catch (e) {
        console.error("Failed to deauthenticate remote", e);
    } finally {
        isBusy.value = false;
    }
}

function replaceRemoteState(remote: RemoteFullRs) {
    const model = {
        id: remote.Id,
        config: new EditableRemoteConfig(toRemoteConfig(remote)),
        credentials: {
            username: remote.Username,
            password: "",
        },
        info: {
            description: remote.Description,
            status: remote.Status,
        },
    }

    const index = remotes.value.findIndex((it) => it.id == remote.Id);
    if (index >= 0) {
        remotes.value.splice(index, 1, model);
    } else {
        remotes.value.push(model);
    }
}

try {
    isBusy.value = true;
    const allRemotes = await remotesApi.getRemotes();

    remotes.value = await Promise.all(allRemotes.map(async (remote) => {
        let remoteState: RemoteFullRs | null = null;
        try {
            remoteState = await remotesApi.getRemote(remote.Id);
        } catch (e) {
            console.log("Failed to get remote info", e);
        }

        return {
            id: remote.Id,
            config: new EditableRemoteConfig(toRemoteConfig(remote)),
            credentials: {
                username: remoteState?.Username ?? "",
                password: "",
            },
            info: {
                description: remoteState?.Description ?? "",
                status: remoteState?.Status ?? "UNKNOWN",
            },
        };
    }));
} catch (e) {
    console.error("Failed to fetch remotes", e);
} finally {
    isBusy.value = false;
}
</script>

<template>
    <ul>
        <li v-for="remote in remotes" :key="remote.id">
            <RemoteEditor :id="remote.id" :config="remote.config" :credentials="remote.credentials" :info="remote.info"
                :allow-config-change="allowConfigChange" :disabled="isBusy" @save="onUpdate(remote.id, remote.config)"
                @delete="onDelete(remote.id)" @authenticate="(credentials) => onAuthenticate(remote.id, credentials)"
                @deauthenticate="onDeauthenticate(remote.id)" />
        </li>
        <li v-if="allowConfigChange">
            <RemoteEditor :id="null" :config="blankRemote.config" :credentials="blankRemote.credentials"
                :info="blankRemote.info" :allow-config-change="allowConfigChange" :disabled="isBusy"
                @save="onCreate(blankRemote.config)" />
        </li>
    </ul>
</template>

<style lang="css" scoped>
li {
    margin-top: 1em;
    margin-bottom: 1em;
}
</style>
