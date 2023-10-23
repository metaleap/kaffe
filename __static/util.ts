export function emoIconDataHref(emoji: string): string {
    return "data:image/svg+xml,&lt;svg xmlns=%22http://www.w3.org/2000/svg%22 viewBox=%220 0 100 100%22&gt;&lt;text y=%22.9em%22 font-size=%2283%22&gt;" + emoji + "&lt;/text&gt;&lt;/svg&gt;"
}
