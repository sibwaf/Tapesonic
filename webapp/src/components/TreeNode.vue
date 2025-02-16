<script setup lang="ts" generic="T">
import type { TreeNode as TreeNodeModel } from '@/util';

defineProps<{
    node: TreeNodeModel<T>,
}>();
</script>

<template>
    <slot :item="node.value"></slot>
    <ul v-if="node.children.length > 0">
        <li v-for="childNode in node.children">
            <TreeNode :node="childNode">
                <!-- @vue-ignore -->
                <template #default="slotProps">
                    <slot :item="slotProps.item"></slot>
                </template>
            </TreeNode>
        </li>
    </ul>
</template>
