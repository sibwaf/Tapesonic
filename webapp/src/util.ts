export default {
    msToTimestamp(ms: number): string {
        const seconds = (ms / 1000) % 60;
        const minutes = (ms / 1000 / 60) % 60;
        const hours = (ms / 1000 / 60 / 60);

        function format(n: number): string {
            const value = Math.floor(n);
            return value < 10 ? `0${value}` : value.toString();
        }

        return `${format(hours)}:${format(minutes)}:${format(seconds)}`;
    },
}
