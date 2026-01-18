<script setup lang="ts">
import api from '@/api';
import UserEditor, { EditableUserPropertiesImpl } from '@/components/UserEditor.vue';
import { ref } from 'vue';
import { useRouter } from 'vue-router';

const router = useRouter();

const isBusy = ref(false);

const user = ref(new EditableUserPropertiesImpl({
    name: "",
    role: "ADMIN",
    password: "",
    passwordConfirmation: "",
    apiKey: null,
}));

async function onSave() {
    try {
        isBusy.value = true;

        const _ = await api.putRootUser({
            Name: user.value.editedValue.name,
            Password: user.value.editedValue.password,
            Role: user.value.editedValue.role,
        });

        await router.push({ name: "home" });
    } catch (e) {
        console.error("Failed to save first user", e);
    } finally {
        isBusy.value = false;
    }
}
</script>

<template>
    <div>
        <h1>First-time setup</h1>
        <h2>Create your first user</h2>
        <UserEditor :disabled="isBusy" :user="user" :allow-role-editing="false" :require-username="true"
            :require-password="true" @save="onSave()" />
    </div>
</template>
