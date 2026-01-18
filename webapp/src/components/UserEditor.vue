<script lang="ts">
export interface UserProperties {
    name: string;
    role: string;
    password: string;
    passwordConfirmation: string;
    apiKey: string | null;
}

export interface EditableUserProperties extends Editable<UserProperties> {
    get isNameEdited(): boolean

    get isPasswordEdited(): boolean

    get isRoleEdited(): boolean
}

export class EditableUserPropertiesImpl implements EditableUserProperties {

    private state: UserProperties;

    private static makeModified(config: UserProperties): UserProperties {
        return JSON.parse(JSON.stringify(config));
    }

    public constructor(private readonly original: UserProperties) {
        this.state = EditableUserPropertiesImpl.makeModified(original);
    }

    public get editedValue(): UserProperties {
        return this.state;
    }

    public get isNameEdited(): boolean {
        return this.original.name != this.editedValue.name;
    }

    public get isPasswordEdited(): boolean {
        return this.original.password != this.editedValue.password;
    }

    public get isRoleEdited(): boolean {
        return this.original.role != this.editedValue.role;
    }

    public get isEdited(): boolean {
        return this.isNameEdited || this.isPasswordEdited || this.isRoleEdited;
    }

    public reset(): void {
        this.state = EditableUserPropertiesImpl.makeModified(this.original);
    }
}

export function userRsToEditableUserProperties(user: UserRs): EditableUserProperties {
    return new EditableUserPropertiesImpl({
        name: user.Name,
        role: user.Role,
        password: "",
        passwordConfirmation: "",
        apiKey: user.ApiKey == "" ? null : user.ApiKey,
    });
}
</script>

<script setup lang="ts">
import type { UserRs } from '@/api';
import type { Editable } from '@/model/Editable';
import { computed, ref } from 'vue';

const props = defineProps<{
    disabled: boolean,
    allowRoleEditing: boolean,
    requireUsername: boolean,
    requirePassword: boolean,
    user: EditableUserProperties,
}>();

const emit = defineEmits<{
    save: [],
    regenerateApiKey: [],
}>();

const isNameValid = computed(() => props.user.editedValue.name.trim().length > 0);
const isPasswordValid = computed(() => props.user.editedValue.password.length > 0 && props.user.editedValue.password == props.user.editedValue.passwordConfirmation);
const isValid = computed(() => {
    if (!isNameValid.value) {
        if (props.requireUsername || props.user.isNameEdited) {
            return false;
        }
    }
    if (!isPasswordValid.value) {
        if (props.requirePassword || props.user.isPasswordEdited) {
            return false;
        }
    }

    return true;
})

const hideApiKey = ref(true);

function onSave() {
    emit("save");
}

function onRegenerateApiKey() {
    emit("regenerateApiKey");
}
</script>

<template>
    <div>
        <input type="text" placeholder="Username" :disabled="disabled" v-model="user.editedValue.name">
        <br>
        <template v-if="allowRoleEditing">
            <select v-model="user.editedValue.role">
                <option value="USER">Regular user</option>
                <option value="ADMIN">Admin</option>
            </select>
            <br>
        </template>
        <input type="password" placeholder="Password" v-model="user.editedValue.password">
        <br>
        <input type="password" placeholder="Confirm password" v-model="user.editedValue.passwordConfirmation">
        <br>
        <button :disabled="disabled || !user.isEdited || !isValid" @click="onSave">Save</button>

        <template v-if="user.editedValue.apiKey != null">
            <h3>API key</h3>

            <div>
                <input :type="hideApiKey ? 'password' : 'text'" :disabled="true" v-model="user.editedValue.apiKey">
                <button @click="hideApiKey = !hideApiKey">{{ hideApiKey ? "Show" : "Hide" }}</button>
                <button :disabled="disabled" @click="onRegenerateApiKey">Regenerate</button>
            </div>
        </template>
    </div>
</template>
