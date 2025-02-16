export interface TreeNode<T> {
    value: T;
    children: TreeNode<T>[];
}

export default {
    msToTimestamp(ms: number): string {
        const sign = Math.sign(ms);
        ms = Math.abs(ms);

        const millis = ms % 1000;
        const seconds = Math.floor((ms / 1000)) % 60;
        const minutes = Math.floor((ms / 1000 / 60)) % 60;
        const hours = Math.floor((ms / 1000 / 60 / 60));

        function format(n: number): string {
            return `${n}`.padStart(2, "0");
        }

        let hhmmss = `${format(hours)}:${format(minutes)}:${format(seconds)}`;
        if (sign < 0) {
            hhmmss = `-${hhmmss}`;
        }

        if (millis > 0) {
            return hhmmss + "." + `${millis}`.padStart(3, "0");
        } else {
            return hhmmss;
        }
    },
    timestampToMs(timestamp: string): number {
        const matches = timestamp.match(/^(\d{1,2}):(\d{1,2}):(\d{1,2})(?:\.(\d{3}))?$/);
        if (matches == null) {
            return 0;
        }

        let millis = 0;
        millis += Number.parseInt(matches[1]) * 60 * 60 * 1000;
        millis += Number.parseInt(matches[2]) * 60 * 1000;
        millis += Number.parseInt(matches[3]) * 1000;

        if (matches.length > 4) {
            millis += Number.parseInt(matches[4]);
        }

        return millis;
    },

    timestampToDate(timestamp: string): string {
        const match = timestamp.match(/^(\d{4}-\d{2}-\d{2})T/);
        if (match == null) {
            return "";
        }

        return `${match[1]}`;
    },

    distinct<T>(items: T[]): T[] {
        const result: T[] = [];
        for (const item of items) {
            if (!result.includes(item)) {
                result.push(item);
            }
        }
        return result;
    },

    areSetsEqual<T>(first: Set<T>, second: Set<T>): boolean {
        if (first.size != second.size) {
            return false;
        }

        for (const item of first) {
            if (!second.has(item)) {
                return false;
            }
        }

        return true;
    }
}
