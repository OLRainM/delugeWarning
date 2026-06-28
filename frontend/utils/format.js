// utils/format.js —— 状态/级别汉化 + 时间格式化

const statusMap = {
	pending_review: '待复核',
	triggered:      '已触发',
	dispatched:     '已派发',
	confirmed:      '已确认',
	handled:        '已处置',
	archived:       '已归档',
	canceled:       '已撤销',
	pending:        '待确认',
	handling:       '处置中',
	finished:       '已完成'
};

const levelMap = {
	blue:   { text: '蓝色', cls: 'level-blue' },
	yellow: { text: '黄色', cls: 'level-yellow' },
	orange: { text: '橙色', cls: 'level-orange' },
	red:    { text: '红色', cls: 'level-red' }
};

function statusText(s) {
	return statusMap[s] || s;
}

// 格式化为 "MM月DD日 HH:mm" 样式
function fmtTime(iso) {
	if (!iso) return '';
	const d = new Date(iso);
	if (isNaN(d)) return '';
	const pad = n => String(n).padStart(2, '0');
	return `${d.getMonth()+1}月${pad(d.getDate())}日 ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

export default { statusMap, levelMap, statusText, fmtTime };
