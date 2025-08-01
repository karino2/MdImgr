import './style.css';
import {ListFiles, CopyUrl, SaveImage, DeleteFile, SelectDirAndNotify, SetTargetDir, SetTemplate} from '../wailsjs/go/main/App';
import Toastify from 'toastify-js';
import 'toastify-js/src/toastify.css';
import copyIcon from './assets/copy.svg'
import deleteIcon from './assets/delete.svg'

const TARGET_DIR_KEY = 'targetDir'
const TEMPLATE_KEY = 'template'
const INITIAL_TEMPLATE = '![images/SomeDir/$1]("images/SomeDir/$1")'

const templateInput = document.getElementById('template-input')

templateInput.addEventListener('input', async (event) => {
    const template = event.target.value
    localStorage.setItem(TEMPLATE_KEY, template)
    await SetTemplate(template)
});

document.getElementById('reset-template-button').addEventListener('click', async () => {
    templateInput.value = INITIAL_TEMPLATE
    localStorage.setItem(TEMPLATE_KEY, INITIAL_TEMPLATE)
    await SetTemplate(INITIAL_TEMPLATE)
});

async function initializeApp() {
    let template = localStorage.getItem(TEMPLATE_KEY)
    if (template === null) {
        template = INITIAL_TEMPLATE
        localStorage.setItem(TEMPLATE_KEY, template)
    }
    templateInput.value = template
    await SetTemplate(template)

    const storedDir = localStorage.getItem(TARGET_DIR_KEY)
    if (storedDir === null) {
        showToast("Please select target dir")
        await SelectDirAndNotify()
    } else {
        await SetTargetDir(storedDir)
    }

}

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
        build.push(`<button onclick="copyUrl('${f}')"><img src="${copyIcon}">copy url</button>`)
        build.push(`<button onclick="deleteFile('${f}')"><img src="${deleteIcon}">delete</button>`)
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

window.deleteFile = (fname) => {
    DeleteFile(fname)
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

async function setNewDir(dir) {
    if (dir) {
        localStorage.setItem(TARGET_DIR_KEY, dir)
        await SetTargetDir(dir)
    }
}

runtime.EventsOn("set-new-dir", (dir) => {
    setNewDir(dir)
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
                const reader = new FileReader()
                reader.onload = async (e) => {
                    const imageDataUrl = e.target.result
                    await SaveImage(imageDataUrl)
                };
                reader.readAsDataURL(blob)
            }
        }
    }
});



initializeApp()
updateImageList()

