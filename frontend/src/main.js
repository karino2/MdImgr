import './style.css';
import {ListFiles, CopyUrl, SaveImage, DeleteFile, SelectDirAndNotify, SetTargetDir, SetTemplate} from '../wailsjs/go/main/App';
import Toastify from 'toastify-js';
import 'toastify-js/src/toastify.css';
import copyIcon from './assets/copy.svg'
import deleteIcon from './assets/delete.svg'

const TARGET_DIR_KEY = 'targetDir'
const TEMPLATE_KEY = 'template'
const HISTORY_KEY = 'history'
const INITIAL_TEMPLATE = '![images/SomeDir/$1]("images/SomeDir/$1")'
const MAX_HISTORY = 20

const templateInput = document.getElementById('template-input')

templateInput.addEventListener('input', async (event) => {
    const template = event.target.value
    localStorage.setItem(TEMPLATE_KEY, template)
    maySaveHistory({dir: localStorage.getItem(TARGET_DIR_KEY), template: template})
    await SetTemplate(template)
});

async function SetNewTemplate(template) {
    templateInput.value = template
    localStorage.setItem(TEMPLATE_KEY, template)
    await SetTemplate(template)
}


/**
 * @typedef HistoryItem
 * @type {object}
 * @property {string} dir
 * @property {string} template
 */

/**
 * @returns {HistoryItem[]}
 */
function loadHistory() {
    let json = localStorage.getItem(HISTORY_KEY)
    if (json === null)
        return []
    return JSON.parse(json)
}

/**
 * @param {HistoryItem[]} historyList
 */
function storeHistory(historyList) {
    let json = JSON.stringify(historyList)
    localStorage.setItem(HISTORY_KEY, json)
}

/**
 * @param {HistoryItem[]} historyList
 * @param {HistoryItem} newItem
 * @returns {HistoryItem[]}
 */
function updateHistoryList(historyList, newItem) {
    let newList = historyList.filter(
        item=> item.dir != newItem.dir
    )
    if (newList.length > MAX_HISTORY)
        newList = newList.slice(0, MAX_HISTORY)
    
    newList.unshift(newItem)
    return newList
}

/**
 * @type {HistoryItem[]}
 */
let g_historyList = []

/**
 * @type {HistoryItem}
 */
let g_lastHistoryItem = null


/**
 * @param {HistoryItem} newItem
 */
function maySaveHistory(newItem) {
    if (g_lastHistoryItem === null || 
        (g_lastHistoryItem.dir != newItem.dir || g_lastHistoryItem.template != newItem.template)) {
        g_historyList = updateHistoryList(g_historyList, newItem)
        storeHistory(g_historyList)

        g_lastHistoryItem = newItem
    }
}

/**
 * @type {HTMLSelectElement}
 */
const historySelect = document.getElementById('history-dialog-select')

const historyDialog = document.getElementById('history-dialog')


/**
 * @param {HistoryItem[]} historyList
 */
function populateHistory(historyList) {
    historySelect.innerHTML = ''
    if (historyList.length == 0) {
        showToast("No history")
        return false
    }
  historyList.forEach( (history, index) => {
    const option = document.createElement('option')
    const dir = history.dir
    option.value = dir
    option.text = dir.split('/').pop() || dir
    option.dataset.index = index
    historySelect.add(option)
  })
  return true
}

function showHistoryDialog() {
    historyDialog.style.display = 'block'
}

function hideHistoryDialog() {
    historyDialog.style.display = 'none'
}

document.getElementById('history-button').addEventListener('click', () => {
    if (!populateHistory(g_historyList))
        return 

    showHistoryDialog()
})

document.getElementById('history-dialog-cancel-button').addEventListener('click', () => {
    hideHistoryDialog()
})

document.getElementById('history-dialog-ok-button').addEventListener('click', async () => {
  const selectedPath = historySelect.value;
  if (selectedPath) {
    const selectedOption = historySelect.options[historySelect.selectedIndex]
    
    const index = selectedOption.dataset.index

    const cur = g_historyList[index]
    await SetNewTemplate(cur.template)
    await setNewDir(cur.dir)
  }
  hideHistoryDialog()
})



async function initializeApp() {
    let template = localStorage.getItem(TEMPLATE_KEY)
    if (template === null) {
        template = INITIAL_TEMPLATE
        localStorage.setItem(TEMPLATE_KEY, template)
    }
    templateInput.value = template
    await SetTemplate(template)

    g_historyList = loadHistory()

    const storedDir = localStorage.getItem(TARGET_DIR_KEY)
    if (storedDir === null) {
        showToast("Please select target dir")
        await SelectDirAndNotify()
    } else {
        g_lastHistoryItem = {dir: storedDir, template: template}
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
        maySaveHistory({dir: dir, template: templateInput.value})
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

