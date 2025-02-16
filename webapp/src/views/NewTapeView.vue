<script setup lang="ts">
import api, { TapeType, type FullSourceRs, type ListThumbnailRs, type Tape, type TrackRs } from '@/api';
import DateEditor from '@/components/DateEditor.vue';
import TapeTrackSearch from '@/components/TapeTrackSearch.vue';
import Thumbnail from '@/components/Thumbnail.vue';
import ThumbnailSelector from '@/components/ThumbnailSelector.vue';
import util from '@/util';
import { computed, ref, watch } from 'vue';
import { useRouter } from 'vue-router';

const router = useRouter();

enum Stage {
    TRACKS,
    METADATA,
    COVER,
    PREVIEW,
}

const isBusy = ref(false);

const stage = ref<Stage>(Stage.TRACKS);
const totalStageCount = computed(() => Object.values(Stage).length / 2);
const goForwardText = computed(() => stage.value < (totalStageCount.value - 1) ? "Next" : "Create");

const canGoBack = computed(() => stage.value > 0);
const canGoForward = computed(() => {
    if (stage.value == Stage.TRACKS && tracks.value.length == 0) {
        return false;
    }

    return true;
});

async function goBack() {
    stage.value -= 1;
}

async function goForward() {
    if (stage.value < (totalStageCount.value - 1)) {
        stage.value += 1;
        return;
    }

    try {
        isBusy.value = true;

        const result = await api.createTape({
            ...tape.value,
            Tracks: tracks.value
        });

        router.push({ name: "tape", params: { tapeId: result.Id } });
    } catch (e) {
        console.error("Failed to save tape", e);
    } finally {
        isBusy.value = false;
    }
}

const tracks = ref<TrackRs[]>([]);
const trackIds = computed(() => tracks.value.map(it => it.Id));
const sourceIds = computed(() => tracks.value.map(it => it.SourceId));
const uniqueSourceIds = computed(() => [...new Set(sourceIds.value)]);

function addTrack(track: TrackRs) {
    tracks.value.push(track);
}

function removeTrack(index: number) {
    tracks.value.splice(index, 1);
}

const tape = ref<Tape>({
    Id: "00000000-0000-0000-0000-000000000000",
    Type: TapeType.Playlist,
    Name: "",
    Artist: "",
    ReleasedAt: null,
    ThumbnailId: null,
    Tracks: [],
});

const metadataGuess = ref<Tape | null>(null);

async function guessAndUpdateMetadata() {
    try {
        isBusy.value = true;

        const trackIdsValue = trackIds.value;
        const guess = await api.guessTapeMetadata(trackIdsValue);

        tape.value = guess;

        metadataGuess.value = guess;
    } catch (e) {
        console.error("Failed to guess tape metadata", e);
    } finally {
        isBusy.value = false;
    }
}

const thumbnails = ref<{ sourceIds: Set<string>, thumbnails: ListThumbnailRs[] }>({ sourceIds: new Set<string>(), thumbnails: [] });
const thumbnailIds = computed(() => {
    const ids = new Set<string>(thumbnails.value.thumbnails.map(it => it.Id));

    const tapeThumbnailId = tape.value.ThumbnailId;
    if (tapeThumbnailId != null) {
        ids.add(tapeThumbnailId);
    }

    return [...ids];
});

async function updateThumbnails() {
    try {
        isBusy.value = true;

        const sourceIdsValue = uniqueSourceIds.value;
        const response = await api.searchThumbnails(sourceIdsValue);

        thumbnails.value.sourceIds = new Set<string>(sourceIdsValue);
        thumbnails.value.thumbnails = response;
    } catch (e) {
        console.error("Failed to fetch thumbnails", e);
    } finally {
        isBusy.value = false;
    }
}

watch(stage, (newStage) => {
    switch (newStage) {
        case Stage.METADATA:
            (async () => {
                const lastAttempt = metadataGuess.value;
                if (lastAttempt != null) {
                    return;
                }
                if (trackIds.value.length == 0) {
                    return;
                }

                await guessAndUpdateMetadata();
            })();
            break;
        case Stage.COVER:
            (async () => {
                const previousSourceIds = thumbnails.value.sourceIds;
                const currentSourceIds = new Set<string>(uniqueSourceIds.value);
                if (util.areSetsEqual(previousSourceIds, currentSourceIds)) {
                    return;
                }

                await updateThumbnails();
            })();
            break;
    }
});
</script>

<template>
    <div>
        <div v-if="stage == Stage.TRACKS">
            <TapeTrackSearch @add-track="addTrack"></TapeTrackSearch>
            <hr>
            <table>
                <thead>
                    <tr>
                        <th></th>
                        <th>Artist</th>
                        <th>Title</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="track, i in tracks" :key="track.Id">
                        <td><button @click="removeTrack(i)">Remove</button></td>
                        <td>{{ track.Artist }}</td>
                        <td>{{ track.Title }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
        <div v-else-if="stage == Stage.METADATA">
            <div>
                <button :disabled="isBusy" @click="guessAndUpdateMetadata">Re-guess</button>
            </div>
            <table>
                <thead></thead>
                <tbody>
                    <tr>
                        <td>
                            <button v-for="option in [TapeType.Album, TapeType.Playlist]"
                                :disabled="isBusy || tape.Type == option" @click="tape.Type = option">
                                {{ option }}
                            </button>
                        </td>
                    </tr>
                    <tr>
                        <td>Name</td>
                        <td>
                            <input :disabled="isBusy" type="text" v-model="tape.Name">
                        </td>
                    </tr>
                    <tr v-if="tape.Type == TapeType.Album">
                        <td>Artist</td>
                        <td>
                            <input :disabled="isBusy" type="text" v-model="tape.Artist">
                        </td>
                    </tr>
                    <tr v-if="tape.Type == TapeType.Album">
                        <td>Release date</td>
                        <td>
                            <DateEditor :disabled="isBusy" v-model="tape.ReleasedAt" />
                            <button :disabled="isBusy || tape.ReleasedAt == null"
                                @click="tape.ReleasedAt = null">Clear</button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
        <div v-else-if="stage == Stage.COVER">
            <ThumbnailSelector :thumbnail-ids="thumbnailIds" size="12em" v-model="tape.ThumbnailId" />
        </div>
        <div v-else-if="stage == Stage.PREVIEW">
            <Thumbnail size="12em" :id="tape.ThumbnailId" />
            <h3>{{ tape.Name }}</h3>
            <h4 v-if="tape.Type == TapeType.Album">by <em>{{ tape.Artist }}</em></h4>
            <ol>
                <li v-for="track in tracks">
                    <span v-if="track.Artist">{{ track.Artist }}</span>
                    <span v-if="track.Artist && track.Title">&ensp;-&ensp;</span>
                    <span v-if="track.Title">{{ track.Title }}</span>
                </li>
            </ol>
        </div>
        <hr>
        <div>
            <button @click="goBack" :disabled="isBusy || !canGoBack">Back</button>
            <span>{{ stage + 1 }}&nbsp;/&nbsp;{{ totalStageCount }}</span>
            <button @click="goForward" :disabled="isBusy || !canGoForward">{{ goForwardText }}</button>
        </div>
    </div>
</template>
