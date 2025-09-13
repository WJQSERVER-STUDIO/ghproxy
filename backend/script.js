document.addEventListener('DOMContentLoaded', function() {
    fetch('/api/stats')
        .then(response => response.json())
        .then(data => {
            const tableBody = document.getElementById('stats-table-body');
            tableBody.innerHTML = ''; // 清空现有内容

            for (const ip in data) {
                const stats = data[ip];
                const row = document.createElement('tr');

                const ipCell = document.createElement('td');
                ipCell.textContent = stats.ip;
                row.appendChild(ipCell);

                const callCountCell = document.createElement('td');
                callCountCell.textContent = stats.call_count;
                row.appendChild(callCountCell);

                const transferredCell = document.createElement('td');
                transferredCell.textContent = stats.total_transferred;
                row.appendChild(transferredCell);

                const lastCalledCell = document.createElement('td');
                lastCalledCell.textContent = new Date(stats.last_called).toLocaleString();
                row.appendChild(lastCalledCell);

                tableBody.appendChild(row);
            }
        })
        .catch(error => {
            console.error('获取统计数据时出错:', error);
            const tableBody = document.getElementById('stats-table-body');
            tableBody.innerHTML = '<tr><td colspan="4" class="text-center">加载统计数据失败</td></tr>';
        });
});
