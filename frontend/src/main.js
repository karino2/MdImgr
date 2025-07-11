import './style.css';
import {ListFiles} from '../wailsjs/go/main/App';

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
        build.push(`<button>copy</button>`)
        build.push(`<button>delete</button>`)
        build.push("</div>")
        build.push("</div>")
    }
    return build.join("\n")
}

async function start() {
    let imgList = document.getElementById('img-list')

    let files = await ListFiles()
    imgList.innerHTML = buildHtml(files)
}

start()

