<template>
	<view class="container">
		<view style="display:flex;gap:16rpx;margin-bottom:16rpx;">
			<view :class="status===''?'btn-primary':'btn-line'" style="flex:1;text-align:center;padding:16rpx;" @tap="filter('')">全部</view>
			<view :class="status==='pending'?'btn-primary':'btn-line'" style="flex:1;text-align:center;padding:16rpx;" @tap="filter('pending')">待确认</view>
			<view :class="status==='handling'?'btn-primary':'btn-line'" style="flex:1;text-align:center;padding:16rpx;" @tap="filter('handling')">处置中</view>
			<view :class="status==='finished'?'btn-primary':'btn-line'" style="flex:1;text-align:center;padding:16rpx;" @tap="filter('finished')">已完成</view>
		</view>

		<view class="card" v-if="tasks.length===0">
			<view class="muted" style="text-align:center;">暂无任务</view>
		</view>

		<view class="card" v-for="item in tasks" :key="item.id" @tap="open(item.id)"
			:style="'border-left:8rpx solid ' + (item.levelColor || '#ccc')">
			<!-- 顶部：级别标签 + 状态 -->
			<view style="display:flex;justify-content:space-between;align-items:center;">
				<view style="display:flex;align-items:center;gap:12rpx;">
					<text class="tag" :class="item.levelCls">{{item.levelText}}预警</text>
					<text style="font-size:24rpx;color:#666;">任务 #{{item.id}}</text>
				</view>
				<text :style="'font-size:26rpx;font-weight:bold;color:'+item.statusColor">{{item.statusText}}</text>
			</view>
			<!-- 预警标题 -->
			<view style="font-size:30rpx;font-weight:bold;margin:14rpx 0 8rpx;">{{item.alertTitle || '加载中…'}}</view>
			<!-- 预警内容摘要 -->
			<view style="font-size:26rpx;color:#333;line-height:1.6;">{{item.alertContent}}</view>
			<!-- 设备 / 派发时间 -->
			<view style="display:flex;justify-content:space-between;margin-top:12rpx;">
				<text class="muted">📍 {{item.deviceName || '设备信息加载中'}}</text>
				<text class="muted">{{item.timeText}}</text>
			</view>
			<!-- 操作提示 -->
			<view style="margin-top:12rpx;color:#1f6feb;font-size:26rpx;">
				{{item.status==='pending'?'⚡ 点击确认接收并开始处置':item.status==='handling'?'📝 点击提交处置结果':'✅ 已完成处置'}}
			</view>
		</view>

		<view class="nav-grid">
			<view class="nav-item" @tap="goAlerts">预警复核</view>
			<view class="nav-item" @tap="goMine">我的</view>
		</view>
	</view>
</template>

<script>
	import api from '@/utils/request';
	import fmt from '@/utils/format';

	const levelColorMap = { blue: '#1f6feb', yellow: '#e3a008', orange: '#f56300', red: '#e5484d' };
	const statusColorMap = { pending: '#e3a008', handling: '#f56300', finished: '#52c41a' };
	let timer = null;

	export default {
		data() {
			return { tasks: [], status: '' };
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
				api.get('/api/v1/tasks', { status: this.status }).then((res) => {
					const items = (res.items || []).map((t) => ({
						...t,
						statusText: fmt.statusText(t.status),
						statusColor: statusColorMap[t.status] || '#666',
						timeText: fmt.fmtTime(t.created_at),
						levelText: '', levelCls: '', levelColor: '#ccc',
						alertTitle: '', alertContent: '', deviceName: ''
					}));
					this.tasks = items;
					// 批量拉取关联预警详情（并发）
					items.forEach((t, idx) => {
						api.get('/api/v1/alerts/' + t.alert_id).then((res) => {
							const a = res.alert || {};
							const li = fmt.levelMap[a.level] || { text: a.level || '', cls: '' };
							this.$set(this.tasks, idx, {
								...this.tasks[idx],
								levelText: li.text,
								levelCls: li.cls,
								levelColor: levelColorMap[a.level] || '#ccc',
								alertTitle: a.title || '',
								alertContent: a.content || '',
								deviceName: a.device_id || ''
							});
						}).catch(() => {});
					});
				});
			},
			open(id) { uni.navigateTo({ url: '/pages/gridworker/task-detail/task-detail?id=' + id }); },
			goAlerts() { uni.navigateTo({ url: '/pages/gridworker/alerts/alerts' }); },
			goMine() { uni.navigateTo({ url: '/pages/gridworker/mine/mine' }); }
		}
	};
</script>
