// Results behind the benchmark charts and tables, updated together. The suite reruns with the commands on the
// Benchmarks page; the comparison comes from a harness that runs the old and new versions side by side.

export const environment = 'Apple M4 Max, Go 1.27.0, median of 5 runs';

/** The two versions the charts compare. */
export const versions = { before: 'v1', after: 'v2' } as const;

/** What golog v2 is compared with: slog's own text and JSON handlers, and v1's console logger. */
export type Group = 'text' | 'json' | 'v1';

/** The names of the two sides of each comparison. v1 always colours, so it is compared with golog in colour. */
export const baselines: Record<Group, { before: string; after: string }> = {
	text: { before: 'slog.TextHandler', after: 'v2' },
	json: { before: 'slog.JSONHandler', after: 'v2' },
	v1: { before: 'v1', after: 'v2 in colour' },
};

export interface Result {
	ns: number;
	allocs: number;
}

export interface Comparison {
	scenario: string;
	before: Result;
	after: Result;
}

/**
 * The same records on the same harness, one record per op, through a *slog.Logger with LogAttrs (v1: its Info with
 * the pairs formatted by fmt.Sprintf), to a writer that drops its input.
 */
export const comparison: Record<Group, Comparison[]> = {
	text: [
		{ scenario: 'Message only', before: { ns: 259.0, allocs: 0 }, after: { ns: 164.5, allocs: 0 } },
		{ scenario: '5 attributes', before: { ns: 446.5, allocs: 0 }, after: { ns: 302.5, allocs: 0 } },
		{ scenario: '10 attributes', before: { ns: 837.0, allocs: 2 }, after: { ns: 508.5, allocs: 1 } },
		{ scenario: 'Groups', before: { ns: 530.7, allocs: 0 }, after: { ns: 334.0, allocs: 0 } },
		{ scenario: '5 attributes, 16 goroutines', before: { ns: 205.6, allocs: 0 }, after: { ns: 168.9, allocs: 0 } },
		{ scenario: 'Below the level', before: { ns: 3.2, allocs: 0 }, after: { ns: 3.6, allocs: 0 } },
	],
	json: [
		{ scenario: 'Message only', before: { ns: 263.6, allocs: 0 }, after: { ns: 164.5, allocs: 0 } },
		{ scenario: '5 attributes', before: { ns: 445.2, allocs: 0 }, after: { ns: 302.5, allocs: 0 } },
		{ scenario: '10 attributes', before: { ns: 777.1, allocs: 3 }, after: { ns: 508.5, allocs: 1 } },
		{ scenario: 'Groups', before: { ns: 483.4, allocs: 0 }, after: { ns: 334.0, allocs: 0 } },
		{ scenario: '5 attributes, 16 goroutines', before: { ns: 208.7, allocs: 0 }, after: { ns: 168.9, allocs: 0 } },
		{ scenario: 'Below the level', before: { ns: 3.2, allocs: 0 }, after: { ns: 3.6, allocs: 0 } },
	],
	v1: [
		{ scenario: 'Message only', before: { ns: 113.3, allocs: 1 }, after: { ns: 174.1, allocs: 0 } },
		{ scenario: '5 attributes', before: { ns: 260.7, allocs: 7 }, after: { ns: 310.4, allocs: 0 } },
		{ scenario: '10 attributes', before: { ns: 490.4, allocs: 10 }, after: { ns: 521.8, allocs: 1 } },
		{ scenario: 'Groups', before: { ns: 314.8, allocs: 7 }, after: { ns: 344.2, allocs: 0 } },
		{ scenario: '5 attributes, 16 goroutines', before: { ns: 204.2, allocs: 7 }, after: { ns: 169.5, allocs: 0 } },
		{ scenario: 'Below the level', before: { ns: 149.4, allocs: 6 }, after: { ns: 3.6, allocs: 0 } },
	],
};

export interface SuiteResult {
	benchmark: string;
	what: string;
	ns: number;
	bytes: number;
	allocs: number;
}

/** v2's own benchmark suite, `go test -bench .`: through a *slog.Logger to io.Discard. */
export const suite: SuiteResult[] = [
	{ benchmark: 'Handler/0 attrs', what: 'A message and nothing else', ns: 165.8, bytes: 0, allocs: 0 },
	{ benchmark: 'Handler/5 attrs', what: 'Strings, ints and a duration', ns: 302.5, bytes: 0, allocs: 0 },
	{ benchmark: 'Handler/10 attrs', what: 'Five more, with an error and a key that needs quoting', ns: 520.5, bytes: 208, allocs: 1 },
	{ benchmark: 'HandlerGroups', what: 'With, WithGroup and a group of five', ns: 332.9, bytes: 0, allocs: 0 },
	{ benchmark: 'HandlerColor', what: 'Five attributes, in colour', ns: 311.8, bytes: 0, allocs: 0 },
	{ benchmark: 'HandlerSource', what: 'Five attributes and the caller’s file and line', ns: 351.7, bytes: 0, allocs: 0 },
	{ benchmark: 'HandlerDisabled', what: 'A Debug record when the level is Info', ns: 4.1, bytes: 0, allocs: 0 },
	{ benchmark: 'HandlerParallel', what: 'Five attributes from 16 goroutines at once', ns: 169.7, bytes: 0, allocs: 0 },
];
