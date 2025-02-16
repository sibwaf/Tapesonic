export interface Editable<T> {

    get editedValue(): T
    get isEdited(): boolean

    reset(): void
}
