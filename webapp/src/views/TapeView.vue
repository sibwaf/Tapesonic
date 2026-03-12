<script setup lang="ts">
import tapes, { type GuessTapeMetadataRs, type TapeFullRs, type TapeRq, TapeType, type TapeRsTrack, type TapeRsArtist } from '@/api/tapes';
import { type TrackRs } from '@/api/tracks';
import thumbnailsApi from '@/api/thumbnails';
import util from '@/util';
import TapeTrackSearch from '@/components/TapeTrackSearch.vue';
import ThumbnailSelector from '@/components/ThumbnailSelector.vue';
import { computed, ref, toRaw, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import DateEditor from '@/components/DateEditor.vue';
import ArtistSelector, { Artist } from '@/components/ArtistSelector.vue';

const router = useRouter();
const route = useRoute();

const tape = ref<TapeFullRs | null>(null);

const name = ref("");
const type = ref(TapeType.Album);
const artist = ref<Artist | null>(null);
const releasedAt = ref<string | null>(null);
const thumbnailId = ref<string | null>(null);
const tracks = ref<TapeRsTrack[]>([]);

const guessedMetadata = ref<GuessTapeMetadataRs | null>(null);
const thumbnailIds = ref<string[]>([]);

const sourceIds = computed(() => tracks.value.map(it => it.SourceId));
const uniqueSourceIds = computed(() => [...new Set<string>(sourceIds.value)].sort());

const trackIds = computed(() => tracks.value.map(it => it.Id));
const trackArtists = computed(() => tracks.value.map(it => it.Artist != null ? new Artist(it.Artist.Id, it.Artist.Name) : null).filter(it => it != null));

const albumTrackIds = computed(() => {
    if (type.value == TapeType.Album) {
        return trackIds.value;
    } else {
        return null;
    }
});

const isEdited = computed(() => {
    const tapeValue = tape.value;
    if (tapeValue == null) {
        return false;
    }

    return name.value != tapeValue.Name
        || type.value != tapeValue.Type
        || artist.value?.id != tapeValue.Artist?.Id
        || releasedAt.value != tapeValue.ReleasedAt
        || thumbnailId.value != tapeValue.ThumbnailId
        || JSON.stringify(trackIds.value) != JSON.stringify(tapeValue.Tracks.map(it => it.Id));
});

const isBusy = ref(false);

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

function onApplyGuessedMetadata(guessedMetadata: GuessTapeMetadataRs) {
    if (guessedMetadata.Artist) {
        artist.value = {
            id: guessedMetadata.Artist.Id,
            name: guessedMetadata.Artist.Name,
        };
    }
    if (guessedMetadata.Name) {
        name.value = guessedMetadata.Name;
    }
    if (guessedMetadata.ReleasedAt) {
        releasedAt.value = guessedMetadata.ReleasedAt;
    }
}

function onReset() {
    const tapeValue = tape.value;
    if (tapeValue == null) {
        return;
    }

    name.value = tapeValue.Name;
    type.value = tapeValue.Type;
    releasedAt.value = tapeValue.ReleasedAt;
    thumbnailId.value = tapeValue.ThumbnailId;
    tracks.value = [...tapeValue.Tracks];

    const tapeArtist = tapeValue.Artist;
    if (tapeArtist != null) {
        artist.value = new Artist(tapeArtist.Id, tapeArtist.Name);
    } else {
        artist.value = null;
    }
}

async function onSave() {
    const tapeValue = tape.value;
    if (tapeValue == null) {
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

        tape.value = await tapes.updateTape(tapeValue.Id, tapeRq);
        onReset();
    } catch (e) {
        console.error("Failed to save tape", e);
    } finally {
        isBusy.value = false;
    }
}

async function onDelete() {
    const tapeValue = tape.value;
    if (tapeValue == null) {
        return;
    }

    try {
        isBusy.value = true;

        await tapes.deleteTape(tapeValue.Id);

        router.push({ path: `/` });
    } catch (e) {
        console.error("Failed to delete tape", e);
    } finally {
        isBusy.value = false;
    }
}

watch(albumTrackIds, async (trackIds) => {
    guessedMetadata.value = null;

    if (trackIds == null || trackIds.length == 0) {
        return;
    }

    try {
        guessedMetadata.value = await tapes.guessTapeMetadata({ TrackIds: trackIds });
    } catch (e) {
        console.error("Failed to guess metadata", e);
    }
}, { immediate: true });

watch(uniqueSourceIds, async (sourceIds) => {
    if (sourceIds.length == 0) {
        thumbnailIds.value = [];
        return;
    }

    try {
        const thumbnails = await thumbnailsApi.searchThumbnails(sourceIds);
        thumbnailIds.value = thumbnails.map(it => it.Id);
    } catch (e) {
        console.error("Failed to fetch thumbnails", e);
    }
}, { immediate: true });

(async () => {
    try {
        isBusy.value = true;

        tape.value = await tapes.getTape(route.params.tapeId as string);
        onReset();
    } catch (e) {
        console.error("Failed to load tape", e);
    } finally {
        isBusy.value = false;
    }
})();
</script>

<template>
    <div v-if="tape == null">
        Loading...
    </div>
    <div v-else>
        <div>
            <button v-for="tapeType in [TapeType.Album, TapeType.Playlist]" :disabled="type == tapeType"
                @click="type = tapeType">
                {{ tapeType }}
            </button>
        </div>
        <table>
            <thead>
                <tr>
                    <th></th>
                    <th></th>
                    <th v-if="type == TapeType.Album">
                        Guessed
                        <button @click="onApplyGuessedMetadata(guessedMetadata!!)"
                            :disabled="guessedMetadata == null">Apply all</button>
                    </th>
                </tr>
            </thead>
            <tbody>
                <tr>
                    <td>Name</td>
                    <td>
                        <input type="text" v-model="name">
                    </td>
                    <td v-if="type == TapeType.Album">
                        <input type="text" disabled="true" :value="guessedMetadata?.Name ?? ''">
                        <button v-if="guessedMetadata?.Name" @click="name = guessedMetadata.Name">Apply</button>
                    </td>
                </tr>
                <tr v-if="type == TapeType.Album">
                    <td>Artist</td>
                    <td>
                        <ArtistSelector v-model="artist" :suggestions="trackArtists" />
                    </td>
                    <td>
                        <input type="text" disabled="true" :value="guessedMetadata?.Artist?.Name ?? ''">
                        <button v-if="guessedMetadata?.Artist"
                            @click="artist = { id: guessedMetadata.Artist.Id, name: guessedMetadata.Artist.Name }">Apply</button>
                    </td>
                </tr>
                <tr v-if="type == TapeType.Album">
                    <td>Release date</td>
                    <td>
                        <DateEditor v-model="releasedAt" />
                        <button @click="releasedAt = null">Clear</button>
                    </td>
                    <td>
                        <input type="date" disabled="true"
                            :value="util.timestampToDate(guessedMetadata?.ReleasedAt ?? '')">
                        <button v-if="guessedMetadata?.ReleasedAt"
                            @click="releasedAt = guessedMetadata.ReleasedAt">Apply</button>
                    </td>
                </tr>
            </tbody>
        </table>
        <div>
            <button :disabled="isBusy || !isEdited" @click="onSave">Save</button>
            <button :disabled="isBusy || !isEdited" @click="onReset">Reset</button>
            <button v-if="tape != null" :disabled="isBusy" @click="onDelete">Delete</button>
        </div>

        <hr>

        <TapeTrackSearch @add-track="onAddTrack($event)" />

        <hr>

        <ThumbnailSelector :thumbnail-ids="thumbnailIds" size="12em" v-model="thumbnailId" />

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
                    <tr v-for="track, index in tracks" :key="track.Id">
                        <td>
                            <button @click="tracks.splice(index, 1)">Remove</button>
                        </td>
                        <td>{{ track.Artist?.Name ?? "" }}</td>
                        <td></td>
                        <td>{{ track.Title }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>
