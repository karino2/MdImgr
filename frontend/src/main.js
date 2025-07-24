import './style.css';
import {ListFiles, CopyUrl, SaveImage} from '../wailsjs/go/main/App';
import Toastify from 'toastify-js';
import 'toastify-js/src/toastify.css';

/**
 * 
 * @param {string[]} files 
 */
function buildHtml(files) {
    let build = []
    for (const f of files) {
        build.push(`<div class="img-container">`)
        build.push(`<div class="img-area">`)
        build.push(`<img src="${f}" class="img-target">`)
        build.push(`</div>`)
        build.push(`<div class="button-area">`)
        build.push(`<button onclick="copyUrl('${f}')"><img src="/src/assets/copy.svg">copy url</button>`)
        build.push(`<button><img src="/src/assets/delete.svg">delete</button>`)
        build.push("</div>")
        build.push("</div>")
    }
    return build.join("\n")
}

let imgListDiv = document.getElementById('img-list')
async function updateImageList() {
    let files = await ListFiles()
    imgListDiv.innerHTML = buildHtml(files)
}

window.copyUrl = (fname) => {
    CopyUrl(fname)
}

runtime.EventsOn("image-list-update", async () => {
    await updateImageList()
})

function showToast(msg) {
    Toastify({
        text: msg
    }).showToast()
}

runtime.EventsOn("show-toast", (msg) => {
    showToast(msg)
})



document.addEventListener('paste', async (event) => {
    console.log("paste called")
    const clipboardData = event.clipboardData || window.clipboardData;

    if (!clipboardData) {
        return;
    }

    for (const item of clipboardData.items) {
        if (item.type.indexOf('image') === 0) {
            const blob = item.getAsFile();

            if (blob) {
                const reader = new FileReader();
                reader.onload = async (e) => {
                    const imageDataUrl = e.target.result
                    await SaveImage(imageDataUrl)
                    showToast("Url copied")
                };
                reader.readAsDataURL(blob)
            }
        }
    }
});



updateImageList()

