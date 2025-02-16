<script setup lang="ts">
import { type Tape, TapeType, type TrackRs } from '@/api';
import api from '@/api';
import util from '@/util';
import TapeTrackSearch from '@/components/TapeTrackSearch.vue';
import ThumbnailSelector from '@/components/ThumbnailSelector.vue';
import { computed, ref, toRaw, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import DateEditor from '@/components/DateEditor.vue';

enum State {
    LOADING,
    LOADING_OK,
    LOADING_ERROR,
    SAVING,
    SAVING_OK,
    SAVING_ERROR,
    DELETING,
    DELETING_OK,
    DELETING_ERROR,
}

const router = useRouter();
const route = useRoute();
const tapeId = computed(() => route.params.tapeId as string);
const isNewTape = computed(() => tapeId.value == "new");

const tape = ref<Tape | null>(null);
const editedTape = ref<Tape | null>(null);

const guessedMetadata = ref<Tape | null>(null);
const thumbnailIds = ref<string[]>([]);

const sourceIds = computed(() => editedTape.value?.Tracks?.map(it => it.SourceId) ?? []);
const uniqueSourceIds = computed(() => [...new Set<string>(sourceIds.value)]);

const trackIds = computed(() => editedTape.value?.Tracks?.map(it => it.Id));

const albumTrackIds = computed(() => {
    const editedTapeValue = editedTape.value;
    if (editedTapeValue == null || editedTapeValue.Type != TapeType.Album) {
        return null;
    }

    return trackIds.value;
});

const state = ref(State.LOADING);

const isEdited = computed(() => JSON.stringify(tape.value) != JSON.stringify(editedTape.value));

const isBusy = computed(() => [State.LOADING, State.SAVING, State.DELETING].includes(state.value));

function onClearReleaseDate(tape: Tape) {
    tape.ReleasedAt = null;
}

function onAddTrack(tape: Tape, track: TrackRs) {
    tape.Tracks.push(track);
}

function onRemoveTrackAt(tape: Tape, index: number) {
    tape.Tracks.splice(index, 1);
}

function onApplyGuessedMetadata(tape: Tape, guessedMetadata: Tape) {
    if (guessedMetadata.Artist) {
        tape.Artist = guessedMetadata.Artist;
    }
    if (guessedMetadata.Name) {
        tape.Name = guessedMetadata.Name;
    }
    if (guessedMetadata.ReleasedAt) {
        tape.ReleasedAt = guessedMetadata.ReleasedAt;
    }
}

function onReset() {
    if (tape.value != null) {
        editedTape.value = structuredClone(toRaw(tape.value));
    } else {
        editedTape.value = {
            Id: "00000000-0000-0000-0000-000000000000",
            Name: "New tape",
            Type: TapeType.Album,
            ThumbnailId: null,
            Tracks: [],
            Artist: "",
            ReleasedAt: null,
        };
    }
}

async function onSave() {
    try {
        state.value = State.SAVING;

        const wasNewTape = isNewTape.value;

        const response = wasNewTape
            ? await api.createTape(editedTape.value!)
            : await api.updateTape(tapeId.value, editedTape.value!);

        tape.value = structuredClone(response);
        onReset();

        if (wasNewTape) {
            router.push({ path: `/tapes/${response.Id}`, force: true });
        }

        state.value = State.SAVING_OK;
    } catch (e) {
        state.value = State.SAVING_ERROR;
        console.error("Failed to save tape", e);
    }
}

async function onDelete() {
    try {
        state.value = State.DELETING;

        await api.deleteTape(tapeId.value);

        router.push({ path: `/` });

        state.value = State.DELETING_OK;
    } catch (e) {
        state.value = State.DELETING_ERROR;
        console.error("Failed to delete tape", e);
    }
}

watch(albumTrackIds, async (trackIds) => {
    guessedMetadata.value = null;

    if (trackIds == null) {
        return;
    }

    try {
        guessedMetadata.value = await api.guessTapeMetadata(trackIds);
    } catch (e) {
        console.error("Failed to guess metadata", e);
    }
});

watch(uniqueSourceIds, async (sourceIds) => {
    // todo: do not trigger if actual ids didn't change

    try {
        const thumbnails = await api.searchThumbnails(sourceIds);
        thumbnailIds.value = thumbnails.map(it => it.Id);
    } catch (e) {
        console.error("Failed to fetch thumbnails", e);
    }
});

if (isNewTape.value) {
    state.value = State.LOADING_OK;
    onReset();
} else {
    (async () => {
        try {
            state.value = State.LOADING;

            tape.value = await api.getTape(tapeId.value);

            state.value = State.LOADING_OK;
        } catch (e) {
            state.value = State.LOADING_ERROR;
            console.error(e);
        }

        onReset();
    })();
}
</script>

<template>
    <template v-if="state == State.LOADING"> Loading... </template>
    <template v-else-if="state == State.LOADING_ERROR"> Failed to load tape {{ tapeId }} </template>
    <template v-else-if="editedTape">
        <div>
            <button v-for="tapeType in [TapeType.Album, TapeType.Playlist]" :disabled="editedTape.Type == tapeType"
                @click="editedTape.Type = tapeType">
                {{ tapeType }}
            </button>
        </div>
        <table>
            <thead>
                <tr>
                    <th></th>
                    <th></th>
                    <th v-if="editedTape.Type == TapeType.Album">
                        Guessed
                        <button @click="onApplyGuessedMetadata(editedTape, guessedMetadata!!)"
                            :disabled="guessedMetadata == null">Apply all</button>
                    </th>
                </tr>
            </thead>
            <tbody>
                <tr>
                    <td>Name</td>
                    <td>
                        <input type="text" v-model="editedTape.Name">
                    </td>
                    <td v-if="editedTape.Type == TapeType.Album">
                        <input type="text" disabled="true" :value="guessedMetadata?.Name ?? ''">
                        <button v-if="guessedMetadata?.Name"
                            @click="editedTape.Name = guessedMetadata.Name">Apply</button>
                    </td>
                </tr>
                <tr v-if="editedTape.Type == TapeType.Album">
                    <td>Artist</td>
                    <td>
                        <input type="text" v-model="editedTape.Artist">
                    </td>
                    <td>
                        <input type="text" disabled="true" :value="guessedMetadata?.Artist ?? ''">
                        <button v-if="guessedMetadata?.Artist"
                            @click="editedTape.Artist = guessedMetadata.Artist">Apply</button>
                    </td>
                </tr>
                <tr v-if="editedTape.Type == TapeType.Album">
                    <td>Release date</td>
                    <td>
                        <DateEditor v-model="editedTape.ReleasedAt"/>
                        <button @click="onClearReleaseDate(editedTape)">Clear</button>
                    </td>
                    <td>
                        <input type="date" disabled="true" :value="util.timestampToDate(guessedMetadata?.ReleasedAt ?? '')">
                        <button v-if="guessedMetadata?.ReleasedAt"
                            @click="editedTape.ReleasedAt = guessedMetadata.ReleasedAt">Apply</button>
                    </td>
                </tr>
            </tbody>
        </table>
        <div>
            <button :disabled="isBusy || !isEdited" @click="onSave">Save</button>
            <button :disabled="isBusy || !isEdited" @click="onReset">Reset</button>
            <button v-if="!isNewTape" :disabled="isBusy" @click="onDelete">Delete</button>

            <div v-if="state == State.SAVING">Saving...</div>
            <div v-else-if="state == State.SAVING_ERROR">Failed to save</div>
            <div v-else-if="state == State.DELETING">Deleting...</div>
            <div v-else-if="state == State.DELETING_ERROR">Failed to delete</div>
        </div>

        <hr>

        <TapeTrackSearch @add-track="onAddTrack(editedTape, $event)" />

        <hr>

        <ThumbnailSelector :thumbnail-ids="thumbnailIds" size="12em" v-model="editedTape.ThumbnailId" />

        <div>
            <table>
                <thead>
                    <tr>
                        <th></th>
                        <th>Artist</th>
                        <th></th>
                        <th>Title</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="track, index in editedTape.Tracks" :key="track.Id">
                        <td>
                            <button @click="onRemoveTrackAt(editedTape, index)">Remove</button>
                        </td>
                        <td>{{ track.Artist }}</td>
                        <td></td>
                        <td>{{ track.Title }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </template>
    <template v-else> Unknown state {{ state }} </template>
</template>
