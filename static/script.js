attachAll(document)

/**
 *
 * @param {ParentNode} root
 */
function attachAll(root) {
    /** @type {NodeListOf<HTMLTextAreaElement>} */
    const elements = root.querySelectorAll('textarea.textarea-auto-grow')
    for (const el of elements) {
        attach(el)
    }
}

/**
 * @param {HTMLTextAreaElement} el
 */
function attach(el) {
    el.addEventListener('input', (e) => {
        el.style.minHeight = ''
        const elStyles = getComputedStyle(el)
        el.style.minHeight = `calc(${el.scrollHeight}px + ${elStyles.borderTopWidth} + ${elStyles.borderBottomWidth})`
    })
}


const observer = new MutationObserver(function(mutations) {
    mutations.forEach(function(mutation) {
        mutation.addedNodes.forEach(function(addedNode) {
            if (addedNode instanceof HTMLElement) {
                attachAll(addedNode)
                return
            }
            console.error("unsupported element", {
                addedNode,
                addedNodeClass: addedNode.constructor.name
            })
            debugger
        });
    });
});
observer.observe(document.body, {childList: true, subtree: true});

