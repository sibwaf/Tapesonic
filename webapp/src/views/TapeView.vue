<script setup lang="ts">
import tapes, { type GuessTapeMetadataRs, type TapeFullRs, type TapeRq, TapeType, type TapeRsTrack, type TapeRsArtist } from '@/api/tapes';
import { type TrackRs } from '@/api/tracks';
import util from '@/util';
import TapeTrackSearch from '@/components/TapeTrackSearch.vue';
import ArtworkSelector from '@/components/ArtworkSelector.vue';
import { computed, ref, toRaw, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import DateEditor from '@/components/DateEditor.vue';
import ArtistSelector, { Artist } from '@/components/ArtistSelector.vue';
import TapeTrackList from '@/components/TapeTrackList.vue';

const router = useRouter();
const route = useRoute();

const tape = ref<TapeFullRs | null>(null);

const name = ref("");
const type = ref(TapeType.Album);
const artist = ref<Artist | null>(null);
const releasedAt = ref<string | null>(null);
const artworkId = ref<string | null>(null);
const tracks = ref<TapeRsTrack[]>([]);

const guessedMetadata = ref<GuessTapeMetadataRs | null>(null);
const artworkIds = computed(() => {
    const result = new Set<string>();

    const tapeArtworkId = tape.value?.ArtworkId ?? null;
    if (tapeArtworkId != null) {
        result.add(tapeArtworkId);
    }

    for (const track of tracks.value) {
        const artworkId = track.ArtworkId;
        if (artworkId != null) {
            result.add(artworkId);
        }
    }

    return [...result];
});

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
        || artworkId.value != tapeValue.ArtworkId
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
        RemoteId: track.RemoteId,
        Artist: artist,
        Title: track.Title,
        ArtworkId: track.ArtworkId,
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
    artworkId.value = tapeValue.ArtworkId;
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
            ArtworkId: artworkId.value,
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

watch(albumTrackIds, async (trackIds, oldTrackIds) => {
    if (trackIds != null && oldTrackIds != null) {
        if (util.areSetsEqual(new Set(trackIds), new Set(oldTrackIds))) {
            return;
        }
    }

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

        <ArtworkSelector :artwork-ids="artworkIds" size="12em" v-model="artworkId" />

        <TapeTrackList v-model="tracks" />
    </div>
</template>
