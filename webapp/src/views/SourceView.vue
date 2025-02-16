<script setup lang="ts">
import api, { type ListSourceHierarchyRs, type FullSourceRs, type TrackRs, type SourceFileRs } from "@/api";
import TrackEditorList from "@/components/TrackEditorList.vue";
import Tree from "@/components/Tree.vue";
import { EditableTrack } from "@/model/EditableTrack";
import { type TreeNode } from "@/util";
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";

enum State {
    LOADING,
    LOADING_ERROR,
    IDLE,
    SAVING_TRACKS,
    DELETING_FILE,
}

const state = ref(State.LOADING);

const route = useRoute();
const sourceId = computed(() => route.params.sourceId as string);

const source = ref<FullSourceRs | null>(null);
const hierarchy = ref<TreeNode<ListSourceHierarchyRs>[]>([]);
const tracks = ref<TrackRs[]>([]);
const file = ref<SourceFileRs | null>(null);

const editedTracks = ref<EditableTrack[]>([]);
const editedTrackSourceIds = computed(() => new Set(editedTracks.value.filter(it => it.isEdited).map(it => it.editedValue.SourceId)));

function buildHierarchyTree(items: ListSourceHierarchyRs[]): TreeNode<ListSourceHierarchyRs>[] {
    const lookup = new Map<string, TreeNode<ListSourceHierarchyRs>>();
    const roots: TreeNode<ListSourceHierarchyRs>[] = [];

    const remaining = [...items];
    let lastRemainingCount = remaining.length + 1;

    while (remaining.length > 0 && lastRemainingCount > remaining.length) {
        lastRemainingCount = remaining.length;

        for (let i = 0; i < remaining.length; i++) {
            const item = remaining[i];

            let lookupNode = lookup.get(item.Id);
            if (lookupNode == null) {
                lookupNode = { value: item, children: [] };
                lookup.set(item.Id, lookupNode);
            }

            if (item.ParentId == null) {
                roots.push(lookupNode);
                remaining.splice(i, 1);
                i -= 1;
                continue;
            }

            const parentNode = lookup.get(item.ParentId);
            if (parentNode == null) {
                continue;
            }

            parentNode.children.push(lookupNode);
            remaining.splice(i, 1);
            i -= 1;
        }
    }

    for (const [_, node] of lookup.entries()) {
        node.children.sort((n1, n2) => n1.value.ListIndex - n2.value.ListIndex);
    }

    return roots;
}

const orderedSourceHierarchy = computed(() => {
    function addAll(node: TreeNode<ListSourceHierarchyRs>, collector: ListSourceHierarchyRs[]) {
        if (collector.includes(node.value)) {
            return;
        }

        collector.push(node.value);
        for (const child of node.children) {
            addAll(child, collector);
        }
    }

    const result: ListSourceHierarchyRs[] = [];
    for (const node of hierarchy.value) {
        addAll(node, result)
    }
    return result;
});

async function saveTracks() {
    try {
        const tracksValue = tracks.value;

        for (const sourceId of editedTrackSourceIds.value) {
            const request = editedTracks.value.map(it => it.editedValue).filter(it => it.SourceId == sourceId);
            const response = await api.replaceSourceTracks(sourceId, request);

            for (let i = 0; i < tracksValue.length; i++) {
                if (tracksValue[i].SourceId == sourceId) {
                    tracksValue.splice(i, 1);
                    i -= 1;
                }
            }
            tracksValue.push(...response);
        }

        // todo: will get out-of-sync if any of the requests after the first one fail, needs a proper batch api
        resetTracks();
    } catch (e) {
        console.error(e);
    } finally {
        if (state.value == State.SAVING_TRACKS) {
            state.value = State.IDLE;
        }
    }
}

function resetTracks() {
    editedTracks.value.splice(0, editedTracks.value.length, ...tracks.value.map(it => new EditableTrack(it)));
}

async function deleteFile() {
    const sourceIdValue = sourceId.value;
    try {
        await api.deleteSourceFile(sourceIdValue);

        if (sourceIdValue != sourceId.value) {
            return;
        }

        file.value = null;
    } catch (e) {
        console.error(e);
    } finally {
        if (state.value == State.DELETING_FILE) {
            state.value = State.IDLE;
        }
    }
}

watch(sourceId, (newSourceId) => {
    (async () => {
        try {
            state.value = State.LOADING;

            const sourceAsync = api.getSource(newSourceId);
            const hierarchyAsync = api.getSourceHierarchy(newSourceId);
            const tracksAsync = api.getSourceTracks(newSourceId, true);
            const fileAsync = api.getSourceFile(newSourceId);

            const sourceResult = await sourceAsync;
            const hierarchyResult = buildHierarchyTree(await hierarchyAsync);
            const tracksResult = await tracksAsync;
            const fileResult = await fileAsync;

            if (sourceId.value != newSourceId) {
                return;
            }

            source.value = sourceResult;
            hierarchy.value = hierarchyResult;
            tracks.value = tracksResult;
            file.value = fileResult;

            resetTracks();

            state.value = State.IDLE;
        } catch (e) {
            console.error(e);

            if (sourceId.value != newSourceId) {
                return;
            }
            state.value = State.LOADING_ERROR;
        }
    })();
}, { immediate: true });
</script>

<template>
    <template v-if="state == State.LOADING"> Loading... </template>
    <template v-else-if="state == State.LOADING_ERROR"> Failed to load source {{ sourceId }} </template>
    <template v-else-if="source != null">
        <Tree :roots="hierarchy">
            <template #default="{ item }">
                <span v-if="item.Id == sourceId">[current] {{ item.Title }}</span>
                <RouterLink v-else :to="`/sources/${item.Id}`">{{ item.Title }}</RouterLink>
            </template>
        </Tree>
        <hr>
        <div>
            <a :href="source.Url" target="_blank">Original URL</a>
        </div>
        <div v-if="source.DurationMs > 0">
            <template v-if="file">
                Downloaded: {{ file.Codec }}
                <button @click="deleteFile" :disabled="state != State.IDLE">Delete media</button>
            </template>
            <template v-else>
                Download is pending
            </template>
        </div>
        <h2>Tracks</h2>
        <div>
            <button @click="saveTracks" :disabled="state != State.IDLE || editedTrackSourceIds.size == 0">
                Save
            </button>
            <button @click="resetTracks" :disabled="state != State.IDLE || editedTrackSourceIds.size == 0">
                Reset
            </button>
        </div>
        <div v-if="state == State.SAVING_TRACKS">Saving...</div>
        <TrackEditorList :modelValue="editedTracks" :orderedSourceHierarchy="orderedSourceHierarchy" />
    </template>
</template>
