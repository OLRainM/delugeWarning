<template>
	<view class="container">
		<view style="display:flex;gap:16rpx;margin-bottom:16rpx;">
			<view :class="status===''?'btn-primary':'btn-line'" style="flex:1;text-align:center;padding:16rpx;" @tap="filter('')">全部</view>
			<view :class="status==='pending'?'btn-primary':'btn-line'" style="flex:1;text-align:center;padding:16rpx;" @tap="filter('pending')">待确认</view>
			<view :class="status==='handling'?'btn-primary':'btn-line'" style="flex:1;text-align:center;padding:16rpx;" @tap="filter('handling')">处置中</view>
			<view :class="status==='finished'?'btn-primary':'btn-line'" style="flex:1;text-align:center;padding:16rpx;" @tap="filter('finished')">已完成</view>
		</view>

		<view class="card" v-if="tasks.length===0"><view class="muted" style="text-align:center;">暂无任务</view></view>
		<view class="card" v-for="item in tasks" :key="item.id" @tap="open(item.id)">
			<view style="display:flex;justify-content:space-between;">
				<text>任务 #{{item.id}}（预警 #{{item.alert_id}}）</text>
				<text class="muted">{{item.status}}</text>
			</view>
			<view class="muted" style="margin-top:8rpx;">{{item.handle_remark || '点击处理'}}</view>
		</view>

		<view class="nav-grid">
			<view class="nav-item" @tap="goAlerts">预警复核</view>
			<view class="nav-item" @tap="goMine">我的</view>
		</view>
	</view>
</template>

<script>
	import api from '@/utils/request';
	export default {
		data() {
			return { tasks: [], status: '' };
		},
		onShow() { this.load(); },
		methods: {
			filter(s) { this.status = s; this.load(); },
			load() {
				api.get('/api/v1/tasks', { status: this.status }).then((res) => {
					this.tasks = res.items || [];
				});
			},
			open(id) { uni.navigateTo({ url: '/pages/gridworker/task-detail/task-detail?id=' + id }); },
			goAlerts() { uni.navigateTo({ url: '/pages/gridworker/alerts/alerts' }); },
			goMine() { uni.navigateTo({ url: '/pages/gridworker/mine/mine' }); }
		}
	};
</script>
