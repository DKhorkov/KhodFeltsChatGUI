// Matches http(s) URLs up to whitespace. Trailing punctuation is trimmed below.
const URL_RE = /https?:\/\/[^\s]+/g
const TRAILING_PUNCT = /[.,;:!?)\]}'"»›]+$/

const HTML_ESCAPES = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;',
}

function escapeHtml(text) {
    return text.replace(/[&<>"']/g, (ch) => HTML_ESCAPES[ch])
}

// linkifyToHtml returns a safe HTML string: raw text is HTML-escaped, then
// http(s) URLs are wrapped in <a class="message-link" data-url="…">. Meant
// to be rendered with v-html; the click handler on the wrapper container
// intercepts .message-link clicks and opens URLs through Wails runtime.
export function linkifyToHtml(text) {
    if (!text) {
        return ''
    }

    let result = ''
    let lastIndex = 0
    URL_RE.lastIndex = 0

    let match
    while ((match = URL_RE.exec(text)) !== null) {
        let url = match[0]
        let trail = ''

        const trailMatch = url.match(TRAILING_PUNCT)
        if (trailMatch) {
            trail = trailMatch[0]
            url = url.slice(0, url.length - trail.length)
        }

        const start = match.index

        if (start > lastIndex) {
            result += escapeHtml(text.slice(lastIndex, start))
        }

        const safeUrl = escapeHtml(url)
        result += `<a class="message-link" data-url="${safeUrl}">${safeUrl}</a>`

        if (trail) {
            result += escapeHtml(trail)
        }

        lastIndex = start + match[0].length
    }

    if (lastIndex < text.length) {
        result += escapeHtml(text.slice(lastIndex))
    }

    return result
}
