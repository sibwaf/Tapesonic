<script setup lang="ts">
import tapes, { TapeType, type GuessTapeMetadataRs, type TapeRq, type TapeRsArtist, type TapeRsTrack } from '@/api/tapes';
import { type TrackRs } from '@/api/tracks';
import thumbnailsApi, { type ListThumbnailRs } from '@/api/thumbnails';
import DateEditor from '@/components/DateEditor.vue';
import TapeTrackSearch from '@/components/TapeTrackSearch.vue';
import Thumbnail from '@/components/Thumbnail.vue';
import ThumbnailSelector from '@/components/ThumbnailSelector.vue';
import util from '@/util';
import { computed, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import ArtistSelector, { type Artist } from '@/components/ArtistSelector.vue';

const router = useRouter();

enum Stage {
    TRACKS,
    METADATA,
    COVER,
    PREVIEW,
}

const isBusy = ref(false);

const name = ref("");
const type = ref(TapeType.Playlist);
const artist = ref<Artist | null>(null);
const releasedAt = ref<string | null>(null);
const thumbnailId = ref<string | null>(null);
const tracks = ref<TapeRsTrack[]>([]);

const trackIds = computed(() => tracks.value.map(it => it.Id));
const sourceIds = computed(() => tracks.value.map(it => it.SourceId));
const uniqueSourceIds = computed(() => [...new Set(sourceIds.value)]);

function onAddTrack(track: TrackRs) {
    let artist: TapeRsArtist | null = null;
    if (track.Artist != null) {
        artist = {
            Id: track.Artist.Id,
            Name: track.Artist.Name,
        };
    }

    tracks.value.push({
        Id: track.Id,
        SourceId: track.SourceId,
        Artist: artist,
        Title: track.Title,
        StartOffsetMs: -1,
        EndOffsetMs: -1,
    });
}

const metadataGuess = ref<GuessTapeMetadataRs | null>(null);

async function guessAndUpdateMetadata() {
    try {
        isBusy.value = true;

        const trackIdsValue = trackIds.value;
        const guess = await tapes.guessTapeMetadata({
            TrackIds: trackIdsValue,
        });

        name.value = guess.Name;
        type.value = guess.Type;
        releasedAt.value = guess.ReleasedAt;
        thumbnailId.value = guess.ThumbnailId;

        if (guess.Artist == null) {
            artist.value = null
        } else {
            artist.value = {
                id: guess.Artist.Id,
                name: guess.Artist.Name,
            };
        }

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

    const tapeThumbnailId = thumbnailId.value;
    if (tapeThumbnailId != null) {
        ids.add(tapeThumbnailId);
    }

    return [...ids];
});

async function updateThumbnails() {
    try {
        isBusy.value = true;

        const sourceIdsValue = uniqueSourceIds.value;
        const response = await thumbnailsApi.searchThumbnails(sourceIdsValue);

        thumbnails.value.sourceIds = new Set<string>(sourceIdsValue);
        thumbnails.value.thumbnails = response;
    } catch (e) {
        console.error("Failed to fetch thumbnails", e);
    } finally {
        isBusy.value = false;
    }
}

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

        const tapeRq: TapeRq = {
            Name: name.value,
            Type: type.value,
            ThumbnailId: thumbnailId.value,
            ArtistId: artist.value?.id ?? null,
            ReleasedAt: releasedAt.value,
            TrackIds: tracks.value.map(it => it.Id),
        };

        const result = await tapes.createTape(tapeRq);

        router.push({ name: "tape", params: { tapeId: result.Id } });
    } catch (e) {
        console.error("Failed to save tape", e);
    } finally {
        isBusy.value = false;
    }
}

watch(stage, async (newStage) => {
    switch (newStage) {
        case Stage.METADATA:
            const lastAttempt = metadataGuess.value;
            if (lastAttempt != null) {
                return;
            }
            if (trackIds.value.length == 0) {
                return;
            }

            await guessAndUpdateMetadata();
            break;
        case Stage.COVER:
            const previousSourceIds = thumbnails.value.sourceIds;
            const currentSourceIds = new Set<string>(uniqueSourceIds.value);
            if (util.areSetsEqual(previousSourceIds, currentSourceIds)) {
                return;
            }

            await updateThumbnails();
            break;
    }
}, { immediate: true });
</script>

<template>
    <div>
        <div v-if="stage == Stage.TRACKS">
            <TapeTrackSearch @add-track="onAddTrack"></TapeTrackSearch>
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
                    <tr v-for="track, index in tracks" :key="track.Id">
                        <td><button @click="tracks.splice(index, 1)">Remove</button></td>
                        <td>{{ track.Artist?.Name ?? "" }}</td>
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
                                :disabled="isBusy || type == option" @click="type = option">
                                {{ option }}
                            </button>
                        </td>
                    </tr>
                    <tr>
                        <td>Name</td>
                        <td>
                            <input :disabled="isBusy" type="text" v-model="name">
                        </td>
                    </tr>
                    <tr v-if="type == TapeType.Album">
                        <td>Artist</td>
                        <td>
                            <ArtistSelector v-model="artist" />
                        </td>
                    </tr>
                    <tr v-if="type == TapeType.Album">
                        <td>Release date</td>
                        <td>
                            <DateEditor :disabled="isBusy" v-model="releasedAt" />
                            <button :disabled="isBusy || releasedAt == null" @click="releasedAt = null">Clear</button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
        <div v-else-if="stage == Stage.COVER">
            <ThumbnailSelector :thumbnail-ids="thumbnailIds" size="12em" v-model="thumbnailId" />
        </div>
        <div v-else-if="stage == Stage.PREVIEW">
            <Thumbnail size="12em" :id="thumbnailId" />
            <h3>{{ name }}</h3>
            <h4 v-if="artist">by <em>{{ artist.name }}</em></h4>
            <ol>
                <li v-for="track in tracks">
                    <span v-if="track.Artist">{{ track.Artist.Name }}</span>
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
