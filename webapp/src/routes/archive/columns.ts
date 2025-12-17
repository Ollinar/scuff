import { renderComponent } from "$lib/components/ui/data-table";
import type { Archive } from "$lib/data/model";
import { formatBytes, formatTimestamp } from "$lib/data/utils";
import type { ColumnDef } from "@tanstack/table-core";
import DataTableSortbutton from "./data-table-sort-button.svelte";
import  DataTableActions  from "./data-table-actions.svelte";


export const columns:ColumnDef<Archive>[] = [
    {
        accessorKey:"id",
        header:"ID",
        minSize:50,
        size:50,
    }, {
        accessorKey:"path",
        header:({column})=>{
            return renderComponent(DataTableSortbutton,{
                compProp:{
                    onclick:column.getToggleSortingHandler()
                },
                label:"Path",
            })
        },
        minSize:600,
        size:600,
    }, {
        accessorKey:"size",
        header:({column})=>{
            return renderComponent(DataTableSortbutton,{
                compProp:{
                    onclick:column.getToggleSortingHandler()
                },
                label:"Size",
            })
        },
        cell:({row,cell})=> {
            return formatBytes(row.getValue("size"))
        },
        minSize:100,
        size:100,
    }, {
        accessorKey:"modtime",
                header:({column})=>{
            return renderComponent(DataTableSortbutton,{
                compProp:{
                    onclick:column.getToggleSortingHandler()
                },
                label:"Modified",
            })
        },
        cell:({row})=>{
            return formatTimestamp(row.getValue("modtime"))
        },
        minSize:175,
        size:175,
    },{
    id: "actions",
    cell: ({ row }) => {
      // You can pass whatever you need from `row.original` to the component
      return renderComponent(DataTableActions, { archive: row.original });
    },
    minSize:50,
    size:50,
  },
]