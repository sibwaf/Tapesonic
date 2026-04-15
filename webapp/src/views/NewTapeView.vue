<script setup lang="ts">
import tapes, { TapeType, type GuessTapeMetadataRs, type TapeRq, type TapeRsArtist, type TapeRsTrack } from '@/api/tapes';
import { type TrackRs } from '@/api/tracks';
import DateEditor from '@/components/DateEditor.vue';
import TapeTrackSearch from '@/components/TapeTrackSearch.vue';
import Artwork from '@/components/Artwork.vue';
import ArtworkSelector from '@/components/ArtworkSelector.vue';
import { computed, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import ArtistSelector, { type Artist } from '@/components/ArtistSelector.vue';
import TapeTrackList from '@/components/TapeTrackList.vue';

const router = useRouter();

enum Stage {
    TRACKS,
    METADATA,
    ARTWORK,
    PREVIEW,
}

const isBusy = ref(false);

const name = ref("");
const type = ref(TapeType.Playlist);
const artist = ref<Artist | null>(null);
const releasedAt = ref<string | null>(null);
const artworkId = ref<string | null>(null);
const tracks = ref<TapeRsTrack[]>([]);

const trackIds = computed(() => tracks.value.map(it => it.Id));

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
        RemoteId: track.RemoteId,
        Artist: artist,
        Title: track.Title,
        ArtworkId: track.ArtworkId,
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
        artworkId.value = guess.ArtworkId;

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

const artworkIds = computed(() => {
    const ids = new Set<string>();

    const tapeArtworkId = artworkId.value;
    if (tapeArtworkId != null) {
        ids.add(tapeArtworkId);
    }

    for (const track of tracks.value) {
        const artworkId = track.ArtworkId;
        if (artworkId != null) {
            ids.add(artworkId);
        }
    }

    return [...ids];
});

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
            ArtworkId: artworkId.value,
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
    }
}, { immediate: true });
</script>

<template>
    <div>
        <div v-if="stage == Stage.TRACKS">
            <TapeTrackSearch @add-track="onAddTrack"></TapeTrackSearch>
            <hr>
            <TapeTrackList v-model="tracks" />
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
        <div v-else-if="stage == Stage.ARTWORK">
            <ArtworkSelector :artwork-ids="artworkIds" size="12em" v-model="artworkId" />
        </div>
        <div v-else-if="stage == Stage.PREVIEW">
            <Artwork size="12em" :id="artworkId" />
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
