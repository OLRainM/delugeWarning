<template>
	<view class="container">
		<view class="card" v-if="alerts.length===0 && !loading">
			<view style="text-align:center;color:#52c41a;font-size:32rpx;">✅ 当前本村暂无生效预警</view>
			<view class="muted" style="text-align:center;margin-top:12rpx;">请保持关注，注意防范</view>
		</view>

		<view class="card" v-for="item in alerts" :key="item.id" @tap="openBroadcast(item.id)">
			<view style="display:flex;justify-content:space-between;align-items:center;">
				<text class="tag" :class="item.levelInfo.cls">{{item.levelInfo.text}}预警</text>
				<text class="muted">{{item.statusText}}</text>
			</view>
			<view style="font-size:32rpx;font-weight:bold;margin:16rpx 0;">{{item.title}}</view>
			<view style="line-height:1.6;">{{item.content}}</view>
			<view class="muted" style="margin-top:12rpx;">
				发布时间：{{item.timeText}} &nbsp;｜&nbsp; 点击收听广播 ▶
			</view>
		</view>

		<view class="nav-grid">
			<navigator url="/pages/villager/guide/guide" class="nav-item">避险指引</navigator>
			<navigator url="/pages/villager/report/report" class="nav-item">隐患上报</navigator>
			<navigator url="/pages/villager/mine/mine" class="nav-item">我的</navigator>
		</view>
	</view>
</template>

<script>
	import api from '@/utils/request';
	import fmt from '@/utils/format';
	let timer = null;
	export default {
		data() {
			return { alerts: [], loading: true };
		},
		onShow() {
			this.load();
			// 无感刷新：每 30 秒自动更新
			timer = setInterval(() => this.load(), 30000);
		},
		onHide() { clearInterval(timer); },
		onUnload() { clearInterval(timer); },
		methods: {
			load() {
				api.get('/api/v1/village/alerts').then((res) => {
					this.alerts = (res.items || []).map((a) => ({
						...a,
						levelInfo: fmt.levelMap[a.level] || { text: a.level, cls: '' },
						statusText: fmt.statusText(a.status),
						timeText: fmt.fmtTime(a.created_at)
					}));
					this.loading = false;
				}).catch(() => { this.loading = false; });
			},
			openBroadcast(id) {
				uni.navigateTo({ url: '/pages/villager/broadcast/broadcast?id=' + id });
			}
		}
	};
</script>
