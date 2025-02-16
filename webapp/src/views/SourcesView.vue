<script setup lang="ts">
import api, { type GetListSourceRs } from "@/api";
import Thumbnail from "@/components/Thumbnail.vue";
import { ref } from "vue";

enum State {
    LOADING,
    LOADING_OK,
    LOADING_ERROR,
}

const state = ref(State.LOADING);

const sources = ref<GetListSourceRs[]>([]);

(async () => {
    try {
        state.value = State.LOADING;

        sources.value = await api.listSources();

        state.value = State.LOADING_OK;
    } catch (e) {
        state.value = State.LOADING_ERROR;
        console.error(e);
    }
})();
</script>

<template>
    <template v-if="state == State.LOADING"> Loading... </template>
    <template v-else-if="state == State.LOADING_ERROR"> Failed to load sources </template>
    <template v-else-if="state == State.LOADING_OK">
        <table>
            <thead>
                <tr>
                    <th></th>
                    <th>Title</th>
                    <th>Uploader</th>
                    <th>Downloaded</th>
                </tr>
            </thead>
            <tbody>
                <tr v-for="{ Source, File } in sources" :key="Source.Id">
                    <td>
                        <Thumbnail size="6em" :id="Source.ThumbnailId" />
                    </td>
                    <td>
                        {{ Source.Title }}
                    </td>
                    <td>
                        {{ Source.Uploader }}
                    </td>
                    <td>
                        {{ Source.DurationMs > 0 ? (File?.Codec ?? "none") : "n/a" }}
                    </td>
                    <td>
                        <RouterLink :to="`/sources/${Source.Id}`">Edit</RouterLink>
                    </td>
                </tr>
            </tbody>
        </table>
    </template>
    <template v-else>
        Unknown state {{ state }}
    </template>
</template>
