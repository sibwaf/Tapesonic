import type { TrackRs } from "@/api";
import type { Editable } from "@/model/Editable";

export class EditableTrack implements Editable<TrackRs> {

    private state: TrackRs;

    private static makeModified(track: TrackRs): TrackRs {
        return JSON.parse(JSON.stringify(track));
    }

    public constructor(private readonly original: TrackRs) {
        this.state = EditableTrack.makeModified(original);
    }

    public get editedValue(): TrackRs {
        return this.state;
    }

    public get isEdited(): boolean {
        return JSON.stringify(this.original) != JSON.stringify(this.editedValue);
    }

    public reset() {
        this.state = EditableTrack.makeModified(this.original);
    }
}
