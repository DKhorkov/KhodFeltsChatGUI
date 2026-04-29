export function debounce(fn, delayMs) {
    let timer = null
    return (...args) => {
        clearTimeout(timer)
        timer = setTimeout(() => fn(...args), delayMs)
    }
}
