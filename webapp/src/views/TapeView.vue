<script setup lang="ts">
import api, { type Tape, type Playlist } from "@/api";
import { useRoute } from "vue-router";
import { computed, ref, toRaw } from "vue";
import TapeTrackListEditor from "@/components/TapeTrackListEditor.vue";
import router from "@/router";
import { v4 as uuid4 } from "uuid";

enum State {
    LOADING,
    LOADING_OK,
    LOADING_ERROR,
    SAVING,
    SAVING_OK,
    SAVING_ERROR,
    CREATING_PLAYLIST,
    CREATING_PLAYLIST_ERROR,
    CREATING_PLAYLIST_OK,
}

const route = useRoute();
const tapeId = route.params.tapeId as string;

const state = ref(State.LOADING);
const tape = ref<Tape | null>(null);
const editedTape = ref<Tape | null>(null);

const isEdited = computed(() => {
    return JSON.stringify(tape.value) != JSON.stringify(editedTape.value);
});

const isBusy = computed(() => {
    switch (state.value) {
        case State.LOADING:
        case State.SAVING:
        case State.CREATING_PLAYLIST:
            return true;
        default:
            return false;
    }
});

function reset() {
    editedTape.value = structuredClone(toRaw(tape.value));
}

async function save() {
    try {
        state.value = State.SAVING;
        await api.saveTape(tapeId, editedTape.value!);
        tape.value = structuredClone(toRaw(editedTape.value));
        state.value = State.SAVING_OK;
    } catch (e) {
        state.value = State.SAVING_ERROR;
        console.error(e);
    }
}

async function createPlaylist() {
    try {
        state.value = State.CREATING_PLAYLIST;

        const tapeValue = tape.value!;
        const playlist: Playlist = {
            Id: uuid4(),
            Name: tapeValue.Name,
            ThumbnailPath: tapeValue.ThumbnailPath,
            Tracks: tapeValue.Tracks.map(it => ({
                Id: uuid4(),
                TapeTrackId: it.Id,
                TapeTrack: undefined!
            }))
        };

        const response = await api.createPlaylist(playlist);
        state.value = State.CREATING_PLAYLIST_OK;

        router.push({ name: "playlist", params: { "playlistId": response.Id } });
    } catch (e) {
        state.value = State.CREATING_PLAYLIST_ERROR;
        console.error(e);
    }
}

function guessArtistAndTitle() {
    const tape = editedTape.value;
    if (tape == null) {
        return;
    }

    for (const track of tape.Tracks) {
        if (track.Artist == "") {
            const parts = track.Title.split(" - ")
            if (parts.length <= 1) {
                continue;
            }

            const cleanup = (text: string) => {
                const timestampRegex = "(?:\\d{1,}:)?\\d{1,2}:\\d{1,2}";
                const fullRegex = new RegExp(`^\\s*(?:${timestampRegex})?\\s*(.+?)\\s*$`);
                return fullRegex.exec(text)?.[1] ?? text;
            };

            track.Artist = cleanup(parts[0]);
            track.Title = cleanup(parts.slice(1).join(" - "));
        }
    }
}

(async () => {
    try {
        state.value = State.LOADING;
        tape.value = await api.getTape(tapeId);
        state.value = State.LOADING_OK;
    } catch (e) {
        state.value = State.LOADING_ERROR;
        console.error(e);
    }

    reset();
})();
</script>

<template>
    <template v-if="state == State.LOADING">
        Loading...
    </template>
    <template v-else-if="state == State.LOADING_ERROR">
        Failed to load tape {{ tapeId }}
    </template>
    <template v-else-if="editedTape">
        <h1>{{ editedTape.Name }}</h1>
        <h2>by {{ editedTape.AuthorName }}</h2>
        <h2>
            <div>
                <button :disabled="isBusy" @click="createPlaylist">Create playlist from this tape</button>
            </div>
            <div v-if="state == State.CREATING_PLAYLIST_ERROR">Failed to created a playlist</div>
        </h2>

        <button :disabled="isBusy" @click="guessArtistAndTitle">Guess artist/title</button>

        <TapeTrackListEditor v-if="editedTape" v-model="editedTape.Tracks" />

        <button :disabled="!isEdited || isBusy" @click="reset">Reset</button>
        <button :disabled="!isEdited || isBusy" @click="save">Save</button>

        <div v-if="state == State.SAVING">Saving...</div>
        <div v-else-if="state == State.SAVING_OK">Saved</div>
        <div v-else-if="state == State.SAVING_ERROR">Failed to save</div>
    </template>
    <template v-else>
        Unknown error
    </template>
</template>
