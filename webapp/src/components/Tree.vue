<script setup lang="ts" generic="T">
import type { TreeNode as TreeNodeModel } from '@/util';
import TreeNode from '@/components/TreeNode.vue';

defineProps<{
    roots: TreeNodeModel<T>[],
}>();
</script>

<template>
    <ul>
        <li v-for="rootNode in roots">
            <TreeNode :node="rootNode">
                <template #default="{ item }">
                    <slot :item="(item as T)"></slot>
                </template>
            </TreeNode>
        </li>
    </ul>
</template>
