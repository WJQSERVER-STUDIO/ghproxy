const githubForm = document.getElementById('github-form');
const githubLinkInput = document.getElementById('githubLinkInput');
const formattedLinkOutput = document.getElementById('formattedLinkOutput');
const output = document.getElementById('output');
const copyButton = document.getElementById('copyButton');
const openButton = document.getElementById('openButton');
const toast = new bootstrap.Toast(document.getElementById('toast'));

function showToast(message) {
    const toastBody = document.querySelector('.toast-body');
    toastBody.textContent = message;
    toast.show();
}

function formatGithubLink(githubLink) {
    const currentHost = window.location.host;
    let formattedLink = "";

    if (githubLink.startsWith("https://github.com/") || githubLink.startsWith("http://github.com/")) {
        formattedLink = window.location.protocol + "//" + currentHost + "/github.com" + githubLink.substring(githubLink.indexOf("/", 8));
    } else if (githubLink.startsWith("github.com/")) {
        formattedLink = window.location.protocol + "//" + currentHost + "/" + githubLink;
    } else if (githubLink.startsWith("https://raw.githubusercontent.com/") || githubLink.startsWith("http://raw.githubusercontent.com/")) {
        formattedLink = window.location.protocol + "//" + currentHost + githubLink.substring(githubLink.indexOf("/", 7));
    } else if (githubLink.startsWith("raw.githubusercontent.com/")) {
        formattedLink = window.location.protocol + "//" + currentHost + "/" + githubLink;
    } else if (githubLink.startsWith("https://gist.githubusercontent.com/") || githubLink.startsWith("http://gist.githubusercontent.com/")) {
        formattedLink = window.location.protocol + "//" + currentHost + "/gist.github.com" + githubLink.substring(githubLink.indexOf("/", 18));
    } else if (githubLink.startsWith("gist.githubusercontent.com/")) {
        formattedLink = window.location.protocol + "//" + currentHost + "/" + githubLink;
    } else {
        showToast('请输入有效的GitHub链接');
        return null;
    }

    return formattedLink;
}

githubForm.addEventListener('submit', function (e) {
    e.preventDefault();
    const formattedLink = formatGithubLink(githubLinkInput.value);
    if (formattedLink) {
        formattedLinkOutput.textContent = formattedLink;
        output.style.display = 'block';
    }
});

copyButton.addEventListener('click', function () {
    navigator.clipboard.writeText(formattedLinkOutput.textContent).then(() => {
        showToast('链接已复制到剪贴板');
    });
});

openButton.addEventListener('click', function () {
    window.open(formattedLinkOutput.textContent, '_blank');
});

function fetchAPI() {
    fetch('/api/size_limit')
        .then(response => response.json())
        .then(data => {
            document.getElementById('sizeLimitDisplay').textContent = `${data.MaxResponseBodySize} MB`;
        });

    fetch('/api/whitelist/status')
        .then(response => response.json())
        .then(data => {
            document.getElementById('whiteListStatus').textContent = data.Whitelist ? '已开启' : '已关闭';
        });

    fetch('/api/blacklist/status')
        .then(response => response.json())
        .then(data => {
            document.getElementById('blackListStatus').textContent = data.Blacklist ? '已开启' : '已关闭';
        });

    fetch('/api/version')
        .then(response => response.json())
        .then(data => {
            document.getElementById('versionBadge').textContent = data.Version;
        });
}

document.addEventListener('DOMContentLoaded', fetchAPI);