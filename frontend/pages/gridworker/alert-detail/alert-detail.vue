<template>
	<view class="container">
		<view class="card">
			<view style="font-weight:bold;font-size:32rpx;">{{alert.title}}</view>
			<view style="margin:16rpx 0;line-height:1.6;">{{alert.content}}</view>
			<view class="muted">状态：{{statusText}} &nbsp;｜&nbsp; 来源：{{sourceText}}</view>
			<view class="muted" style="margin-top:4rpx;">发布时间：{{timeText}}</view>
		</view>

		<view class="card" v-if="alert.status==='pending_review'">
			<view style="font-weight:bold;margin-bottom:16rpx;">复核操作</view>
			<view style="display:flex;gap:16rpx;">
				<view class="btn-primary" style="flex:1;text-align:center;padding:20rpx;" @tap="review('confirm')">确认发布</view>
				<view class="btn-line" style="flex:1;text-align:center;padding:20rpx;" @tap="review('cancel')">撤销误报</view>
			</view>
		</view>

		<button class="btn-primary" v-if="alert.status==='handled'" @tap="archive">归档预警</button>

		<view class="card">
			<view style="font-weight:bold;margin-bottom:12rpx;">流转记录</view>
			<view v-for="item in logs" :key="item.id" style="padding:12rpx 0;border-bottom:1rpx solid #f0f0f0;">
				<view>{{fmtStatus(item.from_status)}} → {{fmtStatus(item.to_status)}}</view>
				<view class="muted">{{item.remark}} &nbsp;{{fmtTime(item.created_at)}}</view>
			</view>
		</view>
	</view>
</template>

<script>
	import api from '@/utils/request';
	import fmt from '@/utils/format';
	export default {
		data() {
			return { id: 0, alert: {}, logs: [], statusText: '', sourceText: '', timeText: '' };
		},
		onLoad(q) { this.id = Number(q.id); this.load(); },
		methods: {
			load() {
				api.get('/api/v1/alerts/' + this.id).then((res) => {
					this.alert = res.alert || {};
					this.logs = res.logs || [];
					this.statusText = fmt.statusText(this.alert.status);
					this.sourceText = this.alert.source === 'sensor' ? '系统自动触发' : '人工发布';
					this.timeText = fmt.fmtTime(this.alert.created_at);
				});
			},
			review(action) {
				api.post('/api/v1/alerts/' + this.id + '/review', { action }).then(() => {
					uni.showToast({ title: '操作成功' });
					this.load();
				});
			},
			archive() {
				api.post('/api/v1/alerts/' + this.id + '/archive', {}).then(() => {
					uni.showToast({ title: '已归档' });
					this.load();
				});
			},
			fmtStatus(s) { return s ? fmt.statusText(s) : '—'; },
			fmtTime(t) { return fmt.fmtTime(t); }
		}
	};
</script>
