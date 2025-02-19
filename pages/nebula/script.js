(() => {
    'use strict';

    // 初始化基础配置
    const currentYear = new Date().getFullYear();
    document.getElementById('currentYear').textContent = currentYear;
    const toast = new bootstrap.Toast('#liveToast');

    // DOM 元素
    const form = document.getElementById('mainForm');
    const input = document.getElementById('inputUrl');
    const output = document.getElementById('output');
    const outputLink = document.getElementById('outputLink');

    // 获取当前域名
    const CURRENT_PROTOCOL = window.location.protocol.replace(':', '');
    const CURRENT_HOST = window.location.host;
    // 替换协议部分
    document.querySelectorAll('code .protocol').forEach(span => {
        span.textContent = CURRENT_PROTOCOL;
    });
    // 替换域名部分
    document.querySelectorAll('code .host').forEach(span => {
        span.textContent = CURRENT_HOST;
    });

    // URL 转换规则
    const URL_RULES = [
        {
            regex: /^(?:https?:\/\/)?(?:www\.)?(github\.com\/.*)/i,
            build: path => `${location.protocol}//${location.host}/${path}`
        },
        {
            regex: /^(?:https?:\/\/)?(raw\.githubusercontent\.com\/.*)/i,
            build: path => `${location.protocol}//${location.host}/${path}`
        },
        {
            regex: /^(?:https?:\/\/)?(gist\.(?:githubusercontent|github)\.com\/.*)/i,
            build: path => `${location.protocol}//${location.host}/${path}`
        }
    ];

    // 核心功能：链接转换
    function transformGitHubURL(url) {
        const cleanURL = url.trim().replace(/^https?:\/\//i, '');
        for (const rule of URL_RULES) {
            const match = cleanURL.match(rule.regex);
            if (match) return rule.build(match[1]);
        }
        return null;
    }

    // 事件处理
    form.addEventListener('submit', e => {
        e.preventDefault();

        if (!input.checkValidity()) {
            input.classList.add('is-invalid');
            showToast('⚠️ 请输入有效的 GitHub 链接');
            return;
        }

        const result = transformGitHubURL(input.value);
        if (!result) {
            showToast('❌ 不支持的链接格式');
            return;
        }

        outputLink.textContent = result;
        output.hidden = false;
        window.scrollTo({ top: output.offsetTop - 100, behavior: 'smooth' });
    });

    document.getElementById('copyBtn').addEventListener('click', async () => {
        try {
            await navigator.clipboard.writeText(outputLink.textContent);
            showToast('✅ 链接已复制');
        } catch {
            showToast('❌ 复制失败');
        }
    });

    document.getElementById('openBtn').addEventListener('click', () => {
        window.open(outputLink.textContent, '_blank', 'noopener,noreferrer');
    });

    // 服务状态监控
    async function loadServiceStatus() {
        try {
            const [size, whitelist, blacklist, version] = await Promise.all([
                fetchJSON('/api/size_limit'),
                fetchJSON('/api/whitelist/status'),
                fetchJSON('/api/blacklist/status'),
                fetchJSON('/api/version')
            ]);

            updateStatus('sizeLimit', `${size.MaxResponseBodySize}MB`);
            updateStatus('whitelistStatus', whitelist.Whitelist ? '已开启' : '已关闭');
            updateStatus('blacklistStatus', blacklist.Blacklist ? '已开启' : '已关闭');
            updateStatus('version', `Version ${version.Version}`);
        } catch {
            showToast('⚠️ 服务状态获取失败');
        }
    }

    async function fetchJSON(url) {
        const response = await fetch(url);
        if (!response.ok) throw new Error('API Error');
        return response.json();
    }

    function updateStatus(elementId, text) {
        const element = document.getElementById(elementId);
        if (element) element.textContent = text;
    }

    // 工具函数
    function showToast(message) {
        const toastBody = document.querySelector('.toast-body');
        toastBody.textContent = message;
        toast.show();
    }

    // 初始化
    input.addEventListener('input', () => {
        input.classList.remove('is-invalid');
        if (output.hidden === false) output.hidden = true;
    });

    loadServiceStatus();
})();