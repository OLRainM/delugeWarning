<template>
	<view class="container">
		<view class="card">
			<view style="font-size:34rpx;font-weight:bold;">{{info.title}}</view>
			<view style="margin-top:20rpx;line-height:1.6;">{{info.content}}</view>
		</view>

		<!-- 语音可播放 -->
		<view v-if="info.tts_ready">
			<button class="btn-primary" @tap="togglePlay">
				{{playing ? '⏸ 暂停广播' : '▶ 播放语音广播'}}
			</button>
		</view>
		<!-- 语音暂不可用（COS 未就绪 / 无语音） -->
		<view v-else-if="info.tts_url" class="card muted" style="text-align:center;">
			🔈 语音广播准备中，请稍后刷新
		</view>
		<view v-else class="card muted" style="text-align:center;">
			🔇 暂无语音广播
		</view>
	</view>
</template>

<script>
	import api from '@/utils/request';
	let audioCtx = null;
	export default {
		data() {
			return { info: {}, playing: false };
		},
		onLoad(q) {
			api.get('/api/v1/alerts/' + q.id + '/broadcast').then((res) => {
				this.info = res;
			});
		},
		methods: {
			togglePlay() {
				const url = this.info.tts_url;
				if (!url || !this.info.tts_ready) {
					uni.showToast({ title: '语音暂不可用', icon: 'none' });
					return;
				}
				if (!audioCtx) {
					audioCtx = uni.createInnerAudioContext();
					audioCtx.onEnded(() => { this.playing = false; });
					audioCtx.onError((e) => {
						this.playing = false;
						uni.showToast({ title: '播放失败，请重试', icon: 'none' });
						console.error('[audio]', e);
						audioCtx.destroy(); audioCtx = null;
					});
				}
				if (this.playing) {
					audioCtx.pause();
					this.playing = false;
				} else {
					audioCtx.src = url;
					audioCtx.play();
					this.playing = true;
				}
			}
		},
		onUnload() {
			if (audioCtx) { audioCtx.destroy(); audioCtx = null; }
		}
	};
</script>
