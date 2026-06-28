// utils/request.js —— 统一封装后端请求，自动带 token（uni-app 版）

import fmt from './format';

function baseURL() {
	return getApp().globalData.baseURL;
}

function request(method, path, data) {
	return new Promise((resolve, reject) => {
		const token = getApp().globalData.token;
		uni.request({
			url: baseURL() + path,
			method,
			data: data || {},
			header: {
				'Content-Type': 'application/json',
				...(token ? { Authorization: 'Bearer ' + token } : {})
			},
			success(res) {
				if (res.statusCode >= 200 && res.statusCode < 300) {
					resolve(res.data);
				} else {
					uni.showToast({ title: (res.data && res.data.error) || '请求失败', icon: 'none' });
					reject(res.data);
				}
			},
			fail(err) {
				uni.showToast({ title: '网络错误', icon: 'none' });
				reject(err);
			}
		});
	});
}

function toQuery(obj) {
	if (!obj) return '';
	const parts = Object.keys(obj)
		.filter((k) => obj[k] !== undefined && obj[k] !== '')
		.map((k) => `${k}=${encodeURIComponent(obj[k])}`);
	return parts.length ? '?' + parts.join('&') : '';
}

export default {
	get: (p, d) => request('GET', p + toQuery(d)),
	post: (p, d) => request('POST', p, d),
	levelMap: fmt.levelMap
};
