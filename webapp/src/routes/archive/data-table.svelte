<script lang="ts" generics="TData, TValue">
	import {
		type ColumnDef,
		type ColumnSizingState,
		type Header,
		type PaginationState,
		type SortingState,
		type ColumnFiltersState,
		getCoreRowModel,
		getPaginationRowModel,
		getFilteredRowModel,
		getSortedRowModel
	} from '@tanstack/table-core';
	import { createSvelteTable, FlexRender } from '$lib/components/ui/data-table/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import * as Pagination from '$lib/components/ui/pagination/index.js';
	import { Input } from '$lib/components/ui/input';

	type DataTableProps<TData, TValue> = {
		columns: ColumnDef<TData, TValue>[];
		data: TData[];
		pageSize?: number;
	};

	let { data, columns, pageSize = 10 }: DataTableProps<TData, TValue> = $props();

	let pagination = $state<PaginationState>({ pageIndex: 0, pageSize: pageSize });
	let sorting = $state<SortingState>([]);
	let columnFilters = $state<ColumnFiltersState>([]);
	let columnSizing = $state<ColumnSizingState>({});

	let page = $derived(pagination.pageIndex + 1);

	const table = createSvelteTable({
		get data() {
			return data;
		},
		columns,

		columnResizeMode: 'onChange',
		getCoreRowModel: getCoreRowModel(),
		getPaginationRowModel: getPaginationRowModel(),
		getSortedRowModel: getSortedRowModel(),
		getFilteredRowModel: getFilteredRowModel(),
		onPaginationChange: (updater) => {
			if (typeof updater === 'function') {
				pagination = updater(pagination);
			} else {
				pagination = updater;
			}
		},
		onSortingChange: (updater) => {
			if (typeof updater === 'function') {
				sorting = updater(sorting);
			} else {
				sorting = updater;
			}
		},
		onColumnFiltersChange: (updater) => {
			if (typeof updater === 'function') {
				columnFilters = updater(columnFilters);
			} else {
				columnFilters = updater;
			}
		},
		onColumnSizingChange: (updater) => {
			if (typeof updater === 'function') {
				columnSizing = updater(columnSizing);
			} else {
				columnSizing = updater;
			}
		},
		state: {
			get pagination() {
				return pagination;
			},
			get sorting() {
				return sorting;
			},
			get columnFilters() {
				return columnFilters;
			},
			get columnSizing() {
				return columnSizing;
			}
		}
	});
</script>

<div>
	<div class="flex items-center py-4">
		<Input
			placeholder="Filter paths..."
			value={(table.getColumn('path')?.getFilterValue() as string) ?? ''}
			onchange={(e) => {
				table.getColumn('path')?.setFilterValue(e.currentTarget.value);
			}}
			oninput={(e) => {
				table.getColumn('path')?.setFilterValue(e.currentTarget.value);
			}}
			class="max-w-sm"
		/>
	</div>
	<div class="px-auto min-w-full rounded-md border">
		<Table.Root class="mx-auto min-w-max rounded-md" style="width:{table.getCenterTotalSize()}px">
			<Table.Header class=" ">
				{#each table.getHeaderGroups() as headerGroup (headerGroup.id)}
					<Table.Row>
						{#each headerGroup.headers as header, index (header.id)}
							<Table.Head
								class="relative border-x-2 "
								colspan={header.colSpan}
								style="width:{header.getSize()}px"
							>
								{#if !header.isPlaceholder}
									<FlexRender
										content={header.column.columnDef.header}
										context={header.getContext()}
									/>

									{@render resizeHandle(header)}
								{/if}
							</Table.Head>
						{/each}
					</Table.Row>
				{/each}
			</Table.Header>
			<Table.Body>
				{#each table.getRowModel().rows as row (row.id)}
					<Table.Row data-state={row.getIsSelected() && 'selected'}>
						{#each row.getVisibleCells() as cell (cell.id)}
							<Table.Cell class="border-x-2" style={`width: ${cell.column.getSize()}px`}>
								<FlexRender content={cell.column.columnDef.cell} context={cell.getContext()} />
							</Table.Cell>
						{/each}
					</Table.Row>
				{:else}
					<Table.Row>
						<Table.Cell colspan={columns.length} class="h-24 text-center">No results.</Table.Cell>
					</Table.Row>
				{/each}
			</Table.Body>
		</Table.Root>
	</div>
	<div class="flex w-full space-x-2 py-4">
		<div class="ml-auto">
			<Pagination.Root bind:page count={table.getRowCount()} perPage={pageSize}>
				{#snippet children({ pages, currentPage })}
					<Pagination.Content>
						<Pagination.Item>
							<Pagination.PrevButton onclick={() => table.previousPage()} />
						</Pagination.Item>
						{#each pages as page (page.key)}
							{#if page.type === 'ellipsis'}
								<Pagination.Item>
									<Pagination.Ellipsis />
								</Pagination.Item>
							{:else}
								<Pagination.Item>
									<Pagination.Link
										onclick={() => table.setPageIndex(page.value - 1)}
										{page}
										isActive={currentPage === page.value}
									>
										{page.value}
									</Pagination.Link>
								</Pagination.Item>
							{/if}
						{/each}
						<Pagination.Item>
							<Pagination.NextButton onclick={() => table.nextPage()} />
						</Pagination.Item>
					</Pagination.Content>
				{/snippet}
			</Pagination.Root>
		</div>
	</div>
</div>
{#snippet resizeHandle(header: Header<TData, unknown>)}
	<div
		role="button"
		tabindex="0"
		ondblclick={() => header.column.resetSize()}
		onmousedown={header.getResizeHandler()}
		ontouchstart={header.getResizeHandler()}
		class="absolute top-0 right-0 h-full w-[5px]
		cursor-col-resize bg-black opacity-0 hover:opacity-100
		active:opacity-100"
	></div>
{/snippet}
