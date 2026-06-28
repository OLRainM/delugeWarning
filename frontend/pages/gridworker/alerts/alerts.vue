<template>
	<view class="container">
		<view style="display:flex;gap:16rpx;margin-bottom:16rpx;">
			<view :class="status===''?'btn-primary':'btn-line'" style="flex:1;text-align:center;padding:16rpx;" @tap="filter('')">全部</view>
			<view :class="status==='pending_review'?'btn-primary':'btn-line'" style="flex:1;text-align:center;padding:16rpx;font-size:24rpx;" @tap="filter('pending_review')">待复核</view>
			<view :class="status==='dispatched'?'btn-primary':'btn-line'" style="flex:1;text-align:center;padding:16rpx;" @tap="filter('dispatched')">已派发</view>
			<view :class="status==='handled'?'btn-primary':'btn-line'" style="flex:1;text-align:center;padding:16rpx;" @tap="filter('handled')">待归档</view>
		</view>

		<view class="card" v-if="alerts.length===0"><view class="muted" style="text-align:center;">暂无预警</view></view>
		<view class="card" v-for="item in alerts" :key="item.id" @tap="open(item.id)">
			<view style="display:flex;justify-content:space-between;align-items:center;">
				<text class="tag" :class="item.levelInfo.cls">{{item.levelInfo.text}}预警</text>
				<text class="muted">{{item.statusText}}</text>
			</view>
			<view style="margin-top:12rpx;">{{item.content}}</view>
			<view class="muted" style="margin-top:8rpx;">
				{{item.sourceText}} &nbsp;｜&nbsp; {{item.timeText}}
			</view>
		</view>
	</view>
</template>

<script>
	import api from '@/utils/request';
	import fmt from '@/utils/format';
	let timer = null;
	export default {
		data() {
			return { alerts: [], status: '' };
		},
		onShow() {
			this.load();
			timer = setInterval(() => this.load(), 30000);
		},
		onHide() { clearInterval(timer); },
		onUnload() { clearInterval(timer); },
		methods: {
			filter(s) { this.status = s; this.load(); },
			load() {
				api.get('/api/v1/alerts', { status: this.status }).then((res) => {
					this.alerts = (res.items || []).map((a) => ({
						...a,
						levelInfo: fmt.levelMap[a.level] || { text: a.level, cls: '' },
						statusText: fmt.statusText(a.status),
						sourceText: a.source === 'sensor' ? '系统自动触发' : '人工发布',
						timeText: fmt.fmtTime(a.created_at)
					}));
				});
			},
			open(id) { uni.navigateTo({ url: '/pages/gridworker/alert-detail/alert-detail?id=' + id }); }
		}
	};
</script>
