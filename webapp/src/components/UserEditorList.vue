<script setup lang="ts">
import usersApi, { type UserRq, type UserRs } from '@/api/users';
import { computed, onMounted, ref } from 'vue';
import UserEditor, { type EditableUserProperties, EditableUserPropertiesImpl, userRsToEditableUserProperties } from './UserEditor.vue';

const props = defineProps<{
    currentUser: UserRs,
}>();

const isBusy = ref(false);

interface Entry {
    id: string;
    properties: EditableUserProperties;
}

function createBlankEntry(): Entry {
    return {
        id: "00000000-0000-0000-0000-000000000000",
        properties: new EditableUserPropertiesImpl({
            name: "",
            role: "USER",
            password: "",
            passwordConfirmation: "",
            apiKey: null,
        }),
    };
}

const blankUser = ref(createBlankEntry());
const users = ref<Entry[]>([]);

const filteredUsers = computed(() => users.value.filter(it => it.id != props.currentUser.Id));

onMounted(async () => {
    try {
        isBusy.value = true;

        const rs = await usersApi.getUsers();
        users.value = rs.map(it => userRsToEntry(it));
    } catch (e) {
        console.log("Failed to load user list", e);
    } finally {
        isBusy.value = false;
    }
});

async function onCreate(user: EditableUserProperties) {
    try {
        isBusy.value = true;

        const rq: UserRq = {
            Name: user.editedValue.name,
            Role: user.editedValue.role,
            Password: user.editedValue.password,
        };

        const createdUser = await usersApi.createUser(rq);

        users.value.push(userRsToEntry(createdUser));
        blankUser.value = createBlankEntry();
    } catch (e) {
        console.log("Failed to create user", e);
    } finally {
        isBusy.value = false;
    }
}

async function onUpdate(id: string, user: EditableUserProperties) {
    try {
        isBusy.value = true;

        const rq: UserRq = {
            Name: user.editedValue.name,
            Role: user.editedValue.role,
            Password: user.editedValue.password,
        };
        if (!user.isPasswordEdited) {
            rq.Password = null;
        }

        const updatedUser = await usersApi.patchUser(id, rq);
        if (updatedUser == null) {
            throw `user ${id} doesn't exist`;
        }

        const index = users.value.findIndex(it => it.id == id);
        if (index >= 0) {
            users.value.splice(index, 1, userRsToEntry(updatedUser));
        } else {
            users.value.push(userRsToEntry(updatedUser));
        }
    } catch (e) {
        console.log("Failed to update user", e);
    } finally {
        isBusy.value = false;
    }
}

function userRsToEntry(user: UserRs): Entry {
    return {
        id: user.Id,
        properties: userRsToEditableUserProperties(user),
    }
}
</script>

<template>
    <ul>
        <li v-for="user in filteredUsers" :key="user.id">
            <UserEditor :disabled="isBusy" :allow-role-editing="true" :require-username="false"
                :require-password="false" :user="user.properties" @save="onUpdate(user.id, user.properties)" />
        </li>
        <li>
            <UserEditor :disabled="isBusy" :allow-role-editing="true" :require-username="true" :require-password="true"
                :user="blankUser.properties" @save="onCreate(blankUser.properties)" />
        </li>
    </ul>
</template>

<style lang="css" scoped>
li {
    margin-top: 1em;
    margin-bottom: 1em;
}
</style>
