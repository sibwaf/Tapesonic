<script setup lang="ts">
import { RouterView } from "vue-router"
</script>

<template>
    <span class="header-link">
        <RouterLink to="/">Home</RouterLink>
    </span>
    <span class="header-link">
        <RouterLink to="/sources">Sources</RouterLink>
    </span>
    <span class="header-link">
        <RouterLink to="/tapes/new">New tape</RouterLink>
    </span>
    <hr>
    <RouterView />
</template>

<style>
.header-link {
    padding: 0.25em;
}
</style>
