<template>
	<view class="container">
		<view style="display:flex;gap:16rpx;margin-bottom:16rpx;">
			<view :class="status===''?'btn-primary':'btn-line'" style="flex:1;text-align:center;padding:16rpx;" @tap="filter('')">全部</view>
			<view :class="status==='pending_review'?'btn-primary':'btn-line'" style="flex:1;text-align:center;padding:16rpx;font-size:24rpx;" @tap="filter('pending_review')">待复核</view>
			<view :class="status==='dispatched'?'btn-primary':'btn-line'" style="flex:1;text-align:center;padding:16rpx;" @tap="filter('dispatched')">已派发</view>
			<view :class="status==='handled'?'btn-primary':'btn-line'" style="flex:1;text-align:center;padding:16rpx;" @tap="filter('handled')">待归档</view>
		</view>

		<view class="card" v-for="item in alerts" :key="item.id" @tap="open(item.id)">
			<view style="display:flex;justify-content:space-between;">
				<text class="tag" :class="item.levelInfo.cls">{{item.levelInfo.text}}</text>
				<text class="muted">{{item.status}}</text>
			</view>
			<view style="margin-top:12rpx;">{{item.content}}</view>
			<view class="muted" style="margin-top:8rpx;">来源：{{item.source==='sensor'?'系统自动':'人工'}}</view>
		</view>
	</view>
</template>

<script>
	import api from '@/utils/request';
	export default {
		data() {
			return { alerts: [], status: '' };
		},
		onShow() { this.load(); },
		methods: {
			filter(s) { this.status = s; this.load(); },
			load() {
				api.get('/api/v1/alerts', { status: this.status }).then((res) => {
					this.alerts = (res.items || []).map((a) => ({
						...a, levelInfo: api.levelMap[a.level] || { text: a.level, cls: '' }
					}));
				});
			},
			open(id) { uni.navigateTo({ url: '/pages/gridworker/alert-detail/alert-detail?id=' + id }); }
		}
	};
</script>
