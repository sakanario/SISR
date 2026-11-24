<script lang="ts">
// GTA V gyro aim settings. Talks to /api/v1/gyroaim (not in the generated OpenAPI
// types yet, so plain fetch is used). Changes apply live and are saved by SISR.
import { onDestroy, onMount } from 'svelte';
import FloatingCard from '../FloatingCard.svelte';
import { apiURL } from '$lib/api/client';
import { log } from '$lib/log';
import { toast } from '$lib/toaster/toaster.svelte';

import IcoClose from '~icons/mdi/close';
import IcoCrosshair from '~icons/mdi/crosshairs-gps';

interface GyroAimSettings {
	enabled: boolean;
	sensitivity: number;
	powerCurve: number;
	minOutput: number;
	deadband: number;
	invertX: boolean;
	invertY: boolean;
}

const { onClose }: { onClose?: () => void } = $props();

let settings = $state<GyroAimSettings>();
let gameConnected = $state(false);
let aiming = $state(false);
let calibrating = $state(false);
let calibrated = $state(false);
let biasPitch = $state(0);
let biasYaw = $state(0);
let saveTimer: ReturnType<typeof setTimeout> | undefined;
let statusTimer: ReturnType<typeof setInterval> | undefined;

const DEFAULTS: GyroAimSettings = {
	enabled: true,
	sensitivity: 3,
	powerCurve: 1,
	minOutput: 0.15,
	deadband: 1,
	invertX: false,
	invertY: false
};

async function load() {
	const res = await fetch(`${apiURL}/api/v1/gyroaim`);
	if (!res.ok) throw new Error(`HTTP ${res.status}`);
	settings = (await res.json()) as GyroAimSettings;
}

async function save() {
	if (!settings) return;
	const { enabled, sensitivity, powerCurve, minOutput, deadband, invertX, invertY } = settings;
	const res = await fetch(`${apiURL}/api/v1/gyroaim`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ enabled, sensitivity, powerCurve, minOutput, deadband, invertX, invertY })
	});
	if (!res.ok) throw new Error(`HTTP ${res.status}: ${await res.text()}`);
}

// Slider drags fire many events; send the last value shortly after the user stops.
function scheduleSave() {
	clearTimeout(saveTimer);
	saveTimer = setTimeout(() => {
		save().catch((e) => {
			log.error('Failed to save gyro aim settings', 'error', e);
			toast({ color: 'firebrick', message: 'Failed to save gyro aim settings' });
		});
	}, 150);
}

async function calibrate() {
	try {
		const res = await fetch(`${apiURL}/api/v1/gyroaim/calibrate`, { method: 'POST' });
		if (!res.ok) throw new Error(`HTTP ${res.status}`);
		calibrating = true;
	} catch (e) {
		log.error('Failed to start gyro calibration', 'error', e);
		toast({ color: 'firebrick', message: 'Failed to start gyro calibration' });
	}
}

async function pollStatus() {
	try {
		const res = await fetch(`${apiURL}/api/v1/gyroaim/status`);
		if (!res.ok) return;
		const s = (await res.json()) as {
			gameConnected: boolean;
			aiming: boolean;
			calibrating: boolean;
			calibrated: boolean;
			biasPitch: number;
			biasYaw: number;
		};
		gameConnected = s.gameConnected;
		aiming = s.aiming;
		calibrating = s.calibrating;
		calibrated = s.calibrated;
		biasPitch = s.biasPitch;
		biasYaw = s.biasYaw;
	} catch {
		gameConnected = false;
		aiming = false;
	}
}

onMount(() => {
	load().catch((e) => {
		log.error('Failed to load gyro aim settings', 'error', e);
		toast({ color: 'firebrick', message: 'Failed to load gyro aim settings' });
	});
	void pollStatus();
	statusTimer = setInterval(() => void pollStatus(), 250);
});

onDestroy(() => {
	clearInterval(statusTimer);
	clearTimeout(saveTimer);
});

const fullStickDps = $derived(settings ? Math.round(360 / settings.sensitivity) : 0);
</script>

<FloatingCard>
	<div id="card-content">
		<div>
			<IcoCrosshair style="width: 1.6em; height: 1.6em;" />
			<h2>GTA V Gyro Aim</h2>
			<button class="plain" onclick={() => onClose?.()}>
				<IcoClose style="width: 1.6em; height: 1.6em;" />
			</button>
		</div>
		<div>
			<div class="status">
				<span class="dot" class:on={gameConnected}></span>
				{#if !gameConnected}
					GTA V not detected or paused (needs GyroAim.asi, Story Mode)
				{:else if aiming}
					<strong>Aiming</strong> - gyro active
				{:else}
					Game connected - gyro idle until you aim
				{/if}
			</div>

			{#if settings}
				<div class="checkbox-wrap">
					<label for="gyro-enabled">Enable gyro aim</label>
					<input
						type="checkbox"
						id="gyro-enabled"
						bind:checked={settings.enabled}
						onchange={scheduleSave} />
				</div>

				<div class="setting">
					<div class="row">
						<label for="gyro-sens">Gyro ° Sensitivity</label>
						<span class="value">{settings.sensitivity.toFixed(2)}x</span>
					</div>
					<input
						type="range"
						id="gyro-sens"
						min="0.1"
						max="10"
						step="0.01"
						bind:value={settings.sensitivity}
						oninput={scheduleSave} />
					<p>Full stick when turning at {fullStickDps}°/s. Higher = faster camera.</p>
				</div>

				<div class="setting">
					<div class="row">
						<label for="gyro-curve">Joystick Power Curve</label>
						<span class="value">{settings.powerCurve.toFixed(1)}</span>
					</div>
					<input
						type="range"
						id="gyro-curve"
						min="0.1"
						max="4"
						step="0.1"
						bind:value={settings.powerCurve}
						oninput={scheduleSave} />
					<p>
						How aggressively the joystick output deflects. 0.1: Extremely Aggressive. 1: Linear. 4:
						Extremely Relaxed.
					</p>
				</div>

				<div class="setting">
					<div class="row">
						<label for="gyro-min">Minimum Stick Output</label>
						<span class="value">{Math.round(settings.minOutput * 100)}%</span>
					</div>
					<input
						type="range"
						id="gyro-min"
						min="0"
						max="0.9"
						step="0.01"
						bind:value={settings.minOutput}
						oninput={scheduleSave} />
					<p>Pushes small movements past GTA's stick deadzone. Raise it if slow aiming feels stuck.</p>
				</div>

				<div class="setting">
					<div class="row">
						<label for="gyro-deadband">Gyro Deadband</label>
						<span class="value">{settings.deadband.toFixed(1)}°/s</span>
					</div>
					<input
						type="range"
						id="gyro-deadband"
						min="0"
						max="10"
						step="0.1"
						bind:value={settings.deadband}
						oninput={scheduleSave} />
					<p>Ignores slower movement than this, to hide hand jitter and drift.</p>
				</div>

				<div class="setting">
					<div class="row">
						<span class="label">Gyro Calibration</span>
						<span class="value">
							{#if calibrated}
								bias {biasPitch.toFixed(2)} / {biasYaw.toFixed(2)}°/s
							{:else}
								not calibrated yet
							{/if}
						</span>
					</div>
					<button onclick={calibrate} disabled={calibrating}>
						{calibrating ? 'Calibrating… keep the controller still' : 'Calibrate gyro (2s)'}
					</button>
					<p>
						Removes the gyro's resting drift (e.g. aim slowly sliding down). Put the controller on a
						table and press Calibrate. SISR also recalibrates automatically whenever the controller is
						held still while you are not aiming.
					</p>
				</div>

				<div class="checkbox-wrap">
					<label for="gyro-inv-x">Invert horizontal</label>
					<input type="checkbox" id="gyro-inv-x" bind:checked={settings.invertX} onchange={scheduleSave} />
				</div>
				<div class="checkbox-wrap">
					<label for="gyro-inv-y">Invert vertical</label>
					<input type="checkbox" id="gyro-inv-y" bind:checked={settings.invertY} onchange={scheduleSave} />
				</div>

				<button
					onclick={() => {
						settings = { ...DEFAULTS };
						scheduleSave();
					}}>Reset to defaults</button>
			{:else}
				<p>Loading…</p>
			{/if}
		</div>
	</div>
</FloatingCard>

<style lang="postcss">
#card-content {
	display: flex;
	flex-direction: column;
	gap: 1em;
	height: 100%;
	min-width: min(28em, 90vw);

	& > :first-child {
		display: grid;
		grid-template-columns: min-content 1fr min-content;
		justify-content: center;
		align-items: center;
		gap: 1em;
		& button {
			padding: 0.5em;
		}
	}
	& > :last-child {
		overflow: auto;
		display: grid;
		gap: 1.25em;
		padding: 1em;
		width: 100%;
	}
}

.status {
	display: flex;
	align-items: center;
	gap: 0.6em;
	color: var(--text-muted);
}

.dot {
	width: 0.7em;
	height: 0.7em;
	border-radius: 50%;
	background: var(--background-neutral-alpha);
	flex: none;
	&.on {
		background: #3fbf6b;
	}
}

.setting {
	display: grid;
	gap: 0.4em;
	& .row {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		gap: 1em;
	}
	& label,
	& .label {
		font-weight: bold;
	}
	& .value {
		font-variant-numeric: tabular-nums;
		color: var(--text-muted);
	}
	& p {
		font-size: 0.85em;
		color: var(--text-muted);
	}
}

input[type='range'] {
	width: 100%;
	padding: 0;
	border: none;
	box-shadow: none;
	accent-color: var(--color-primary);
}

.checkbox-wrap {
	display: grid;
	grid-auto-flow: column;
	gap: 0.5em;
	align-items: center;
	justify-content: start;
}

button {
	width: fit-content;
}
</style>
