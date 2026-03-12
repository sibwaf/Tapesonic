import type { SourceTrackRs } from "@/api/sources";
import type { Editable } from "@/model/Editable";

export class EditableTrack implements Editable<SourceTrackRs> {

    private state: SourceTrackRs;

    private static makeModified(track: SourceTrackRs): SourceTrackRs {
        return JSON.parse(JSON.stringify(track));
    }

    public constructor(private readonly original: SourceTrackRs) {
        this.state = EditableTrack.makeModified(original);
    }

    public get editedValue(): SourceTrackRs {
        return this.state;
    }

    public get isEdited(): boolean {
        return JSON.stringify(this.original) != JSON.stringify(this.editedValue);
    }

    public reset() {
        this.state = EditableTrack.makeModified(this.original);
    }
}
