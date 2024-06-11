<script setup lang="ts">
import api, { type Tape, type Playlist, type PlaylistTrack, type Album, type AlbumTrack, type RelatedItems } from "@/api";
import { useRoute } from "vue-router";
import { computed, ref, toRaw } from "vue";
import TapeTrackListEditor from "@/components/TapeTrackListEditor.vue";
import router from "@/router";
import AlbumGrid from "@/components/AlbumGrid.vue";
import PlaylistGrid from "@/components/PlaylistGrid.vue";

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
    CREATING_ALBUM,
    CREATING_ALBUM_ERROR,
    CREATING_ALBUM_OK,
}

const route = useRoute();
const tapeId = route.params.tapeId as string;

const state = ref(State.LOADING);

const tape = ref<Tape | null>(null);
const editedTape = ref<Tape | null>(null);

const relatedItems = ref<RelatedItems | null>(null);

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

const canCreateAlbum = computed(() => {
    const tapeValue = tape.value;
    if (tapeValue == null) {
        return false;
    }

    const tracks = tapeValue.Files.flatMap(it => it.Tracks);
    if (tracks.length == 0) {
        return false;
    }

    const artists = new Set(tracks.map(it => it.Artist));
    if (artists.size != 1) {
        return false;
    }
    if ([...artists][0].trim() == "") {
        return false;
    }

    return true;
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
            Id: undefined!,
            Name: tapeValue.Name,
            ThumbnailPath: tapeValue.ThumbnailPath,
            Tracks: tapeValue.Files
                .flatMap(it => it.Tracks)
                .map(it => {
                    const track: PlaylistTrack = {
                        Id: undefined!,
                        TapeTrackId: it.Id,
                        TapeTrack: undefined!,
                    };
                    return track;
                }),
        };

        const response = await api.createPlaylist(playlist);
        state.value = State.CREATING_PLAYLIST_OK;

        router.push({ name: "playlist", params: { "playlistId": response.Id } });
    } catch (e) {
        state.value = State.CREATING_PLAYLIST_ERROR;
        console.error(e);
    }
}

async function createAlbum() {
    try {
        state.value = State.CREATING_ALBUM;

        const tapeValue = tape.value!;

        const allTracks = tapeValue.Files.flatMap(it => it.Tracks);

        const album: Album = {
            Id: undefined!,
            Name: tapeValue.Name,
            Artist: allTracks[0].Artist,
            ReleaseDate: undefined!,
            ThumbnailPath: tapeValue.ThumbnailPath,
            Tracks: allTracks.map(it => {
                const track: AlbumTrack = {
                    Id: undefined!,
                    TapeTrackId: it.Id,
                    TapeTrack: undefined!,
                };
                return track;
            }),
        };

        const response = await api.createAlbum(album);
        state.value = State.CREATING_ALBUM_OK;

        router.push({ name: "album", params: { "albumId": response.Id } });
    } catch (e) {
        state.value = State.CREATING_ALBUM_ERROR;
        console.error(e);
    }
}

function guessArtistAndTitle() {
    const tape = editedTape.value;
    if (tape == null) {
        return;
    }

    for (const file of tape.Files) {
        for (const track of file.Tracks) {
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
}

function swapArtistAndTitle() {
    const tape = editedTape.value;
    if (tape == null) {
        return;
    }

    for (const file of tape.Files) {
        for (const track of file.Tracks) {
            const temp = track.Artist;
            track.Artist = track.Title;
            track.Title = temp;
        }
    }
}

(async () => {
    try {
        state.value = State.LOADING;

        const tapeAsync = api.getTape(tapeId);
        const relatedItemsAsync = api.getTapeRelationships(tapeId);

        tape.value = await tapeAsync;
        relatedItems.value = await relatedItemsAsync;

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
        <h2 v-if="canCreateAlbum">
            <div>
                <button :disabled="isBusy" @click="createAlbum">Create album from this tape</button>
            </div>
            <div v-if="state == State.CREATING_ALBUM_ERROR">Failed to created an album</div>
        </h2>

        <button :disabled="isBusy" @click="guessArtistAndTitle">Guess artist/title</button>
        <button :disabled="isBusy" @click="swapArtistAndTitle">Swap artist/title</button>

        <div v-for="file in editedTape.Files">
            <h3>{{ file.Name }}</h3>
            <TapeTrackListEditor v-model="file.Tracks" />
        </div>

        <button :disabled="!isEdited || isBusy" @click="reset">Reset</button>
        <button :disabled="!isEdited || isBusy" @click="save">Save</button>

        <div v-if="state == State.SAVING">Saving...</div>
        <div v-else-if="state == State.SAVING_OK">Saved</div>
        <div v-else-if="state == State.SAVING_ERROR">Failed to save</div>

        <template v-if="relatedItems?.Playlists">
            <hr>

            <h2>Linked playlists</h2>
            <PlaylistGrid v-model="relatedItems.Playlists" />
        </template>

        <template v-if="relatedItems?.Albums">
            <hr>

            <h2>Linked albums</h2>
            <AlbumGrid v-model="relatedItems.Albums" />
        </template>
    </template>
    <template v-else>
        Unknown error
    </template>
</template>
