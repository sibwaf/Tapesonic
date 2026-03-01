<script lang="ts">
export class Artist {
    public readonly id: string;
    public readonly name: string;

    constructor(id: string, name: string) {
        this.id = id;
        this.name = name;
    }
}
</script>

<script setup lang="ts">
import artists from '@/api/artists';
import { AutoComplete, Button } from 'primevue';
import { computed, ref, toRaw, watch } from 'vue';

const model = defineModel<Artist | null>();

const props = defineProps<{
    suggestions?: Artist[],
}>();

const rawValue = ref();

function valueAsArtist(value: any): Artist | null {
    if (value == null || typeof (value) !== 'object') {
        return null;
    }

    if (value.id === undefined) {
        return null;
    }
    if (value.name === undefined) {
        return null;
    }

    return new Artist(value.id, value.name);
}

watch(model, (modelValue) => rawValue.value = modelValue, { immediate: true });

watch(rawValue, (rawValue) => {
    const artist = valueAsArtist(toRaw(rawValue));

    if (artist != null) {
        if (model.value?.id != artist.id) {
            model.value = rawValue;
        }
    } else if (rawValue == "" || rawValue == null) {
        if (model.value != null) {
            model.value = null;
        }
    }
}, { immediate: true })

const lastQuery = ref("");

const suggestions = ref<Artist[]>([]);
const isBusy = ref(false);

async function updateSuggestions(query: string) {
    lastQuery.value = query ?? "";

    try {
        isBusy.value = true;
        suggestions.value = (await artists.listArtists(query, 20, 0)).map(it => new Artist(it.Id, it.Name));
    } finally {
        isBusy.value = false;
    }
}

const allSuggestions = computed(() => {
    const options: Artist[] = [];

    function addOption(artist: Artist) {
        if (options.find(it => it.id == artist.id) !== undefined) {
            return;
        }

        options.push(artist);
    }

    const currentValue = model.value;
    if (currentValue != null) {
        addOption(currentValue);
    }

    for (const suggestion of (props.suggestions ?? [])) {
        addOption(suggestion);
    }
    for (const suggestion of suggestions.value) {
        addOption(suggestion);
    }

    return options;
});

async function createArtist(name: string) {
    try {
        isBusy.value = true;

        const artist = await artists.createArtist({
            Name: name,
            Aliases: [],
            MusicBrainzId: null,
        });

        model.value = new Artist(artist.Id, artist.Name);
    } catch (e) {
        console.error("Failed to create an artist", e);
    } finally {
        isBusy.value = false;
    }
}
</script>

<template>
    <AutoComplete option-label="name" show-clear v-model="rawValue" :loading="isBusy" :suggestions="allSuggestions"
        @complete="event => updateSuggestions(event.query)">

        <template #header>
            <Button fluid icon="pi pi-plus" variant="text" :label="lastQuery" @click="createArtist(lastQuery)" />
        </template>

    </AutoComplete>
</template>
