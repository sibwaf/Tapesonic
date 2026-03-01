<script setup lang="ts">
import artistsApi, { type ArtistFullRs, type ArtistRq } from '@/api/artists';
import { Button, Fieldset, InputText, InputGroup, Toolbar } from 'primevue';
import { computed, ref, toRaw } from 'vue';
import { useRoute } from 'vue-router';

const route = useRoute();
const artistId = computed(() => route.params.artistId as string);

const isBusy = ref(false);

const artist = ref<ArtistFullRs | null>(null);

const name = ref("");
const aliases = ref<string[]>([]);

const isEdited = computed(() => {
    return name.value != artist.value?.Name
        || JSON.stringify(aliases.value) != JSON.stringify(artist.value?.Aliases ?? []);
});

function onReset() {
    const artistValue = toRaw(artist.value);
    if (!artistValue) {
        name.value = "";
        aliases.value = [];
    } else {
        name.value = artistValue.Name;
        aliases.value = structuredClone(artistValue.Aliases ?? []);
    }
}

function setAliasAsMainName(index: number) {
    const tmp = aliases.value[index];
    aliases.value[index] = name.value;
    name.value = tmp;
}

(async () => {
    try {
        isBusy.value = true;
        artist.value = await artistsApi.getArtistById(artistId.value);
        onReset();
    } catch (e) {
        console.error("Failed to load artist", e);
    } finally {
        isBusy.value = false;
    }
})();

async function save() {
    const artistValue = artist.value;
    if (!artistValue) {
        return;
    }

    try {
        isBusy.value = true;

        const rq: ArtistRq = {
            Name: name.value,
            Aliases: aliases.value,
            MusicBrainzId: artistValue.MusicBrainzId,
        };

        artist.value = await artistsApi.putArtistById(artistValue.Id, rq);
        onReset();
    } catch (e) {
        console.error("Failed to save artist", e);
    } finally {
        isBusy.value = false;
    }
}
</script>

<template>
    <div>
        <InputText v-model="name" />

        <Fieldset legend="Aliases">
            <div class="artist-alias-row" v-for="_, i in aliases">
                <InputGroup>
                    <Button :disabled="isBusy" icon="pi pi-trash" variant="text" severity="danger"
                        aria-label="Remove alias" @click="aliases.splice(i, 1)" />
                    <Button :disabled="isBusy" variant="text" severity="contrast" @click="setAliasAsMainName(i)">Set as
                        main</Button>
                </InputGroup>

                <InputText :disabled="isBusy" v-model="aliases[i]" />
            </div>

            <div class="artist-alias-row">
                <Button :disabled="isBusy" icon="pi pi-plus" label="Add alias" variant="text" severity="contrast"
                    @click="aliases.push('')" />
            </div>
        </Fieldset>

        <Toolbar>
            <template #start>
                <Button :disabled="isBusy || !isEdited" icon="pi pi-save" label="Save" variant="text"
                    severity="contrast" @click="save" />
            </template>
        </Toolbar>
    </div>
</template>

<style>
.artist-alias-row {
    margin-top: 0.5em;
    margin-bottom: 0.5em;

    display: flex;
    justify-content: start;
}

.artist-alias-row .p-inputgroup {
    width: auto;
}
</style>
