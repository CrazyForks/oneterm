<template>
  <div class="command-template-management">
    <a-spin :tip="loadTip" :spinning="loading">
      <div class="command-template-management-header">
        <a-space>
          <a-input-search
            v-model="searchValue"
            :placeholder="$t('placeholderSearch')"
            :style="{ width: '250px' }"
            class="ops-input ops-input-radius"
            allow-clear
            @search="updateTableData()"
          />
          <div class="ops-list-batch-action" v-show="!!selectedRowKeys.length">
            <span @click="batchDelete">{{ $t('delete') }}</span>
            <span>{{ $t('selectRows', { rows: selectedRowKeys.length }) }}</span>
          </div>
        </a-space>
        <a-space>
          <a-button type="primary" @click="openModal(null)">{{ $t('create') }}</a-button>
          <a-button @click="updateTableData()">{{ $t('refresh') }}</a-button>
        </a-space>
      </div>
      <ops-table
        size="small"
        ref="opsTable"
        class="ops-stripe-table"
        stripe
        show-overflow
        show-header-overflow
        resizable
        :data="tableData"
        :checkbox-config="{ reserve: true, highlight: true, range: true }"
        :row-config="{ keyField: 'id' }"
        :height="tableHeight"
        :filter-config="{ remote: true }"
        @filter-change="handleFilterChange"
        @checkbox-change="onSelectChange"
        @checkbox-all="onSelectChange"
        @checkbox-range-end="onSelectRangeEnd"
      >
        <vxe-column type="checkbox" width="60px"></vxe-column>
        <vxe-column :title="$t('name')" field="name"></vxe-column>
        <vxe-column :title="$t('description')" field="description"></vxe-column>
        <vxe-column :title="$t('oneterm.commandFilter.commandCount')" field="description">
          <template #default="{row}">
            {{ row.cmd_ids.length }}
          </template>
        </vxe-column>
        <vxe-column
          :title="$t('oneterm.commandFilter.category')"
          field="category"
          :filters="categoryFilters"
          :filter-multiple="false"
        >
          <template #default="{row}">
            {{ $t(row.categoryName) }}
          </template>
        </vxe-column>
        <vxe-column :title="$t('created_at')" width="170">
          <template #default="{row}">
            {{ row.createdTimeText }}
          </template>
        </vxe-column>
        <vxe-column :title="$t('operation')" width="100">
          <template #default="{row}">
            <a-space>
              <a @click="openModal(row)"><ops-icon type="icon-xianxing-edit"/></a>
              <a-popconfirm :title="$t('confirmDelete')" @confirm="deleteCommandTemplate(row)">
                <a style="color:red"><ops-icon type="icon-xianxing-delete"/></a>
              </a-popconfirm>
            </a-space>
          </template>
        </vxe-column>
      </ops-table>
      <div class="command-template-management-pagination">
        <a-pagination
          size="small"
          show-size-changer
          :current="currentPage"
          :total="totalResult"
          :show-total="
            (total, range) =>
              $t('pagination.total', {
                range0: range[0],
                range1: range[1],
                total,
              })
          "
          :page-size="pageSize"
          :default-current="1"
          @change="pageOrSizeChange"
          @showSizeChange="pageOrSizeChange"
        />
      </div>
    </a-spin>
    <CommandTemplateModal ref="commandTemplateModal" @submit="updateTableData()" />
  </div>
</template>

<script>
import moment from 'moment'
import { mapState } from 'vuex'
import { getCommandTemplateList, deleteCommandTemplateById } from '@/modules/oneterm/api/commandTemplate.js'
import { COMMAND_CATEGORY, COMMAND_CATEGORY_NAME } from '../constants.js'

import CommandTemplateModal from './commandTemplateModal.vue'

export default {
  name: 'CommandTemplateManagement',
  components: { CommandTemplateModal },
  data() {
    return {
      searchValue: '',
      currentCategory: [],

      tableData: [],
      currentPage: 1,
      pageSize: 20,
      totalResult: 0,
      selectedRowKeys: [],
      loading: false,
      loadTip: '',
    }
  },
  computed: {
    ...mapState({
      windowHeight: (state) => state.windowHeight,
    }),
    tableHeight() {
      return this.windowHeight - 254
    },
    categoryFilters() {
      return Object.values(COMMAND_CATEGORY).map((value) => {
        return {
          value,
          label: this.$t(COMMAND_CATEGORY_NAME[value])
        }
      })
    },
  },
  mounted() {
    this.updateTableData()
  },
  methods: {
    updateTableData() {
      this.loading = true
      const category = this?.currentCategory?.length ? this.currentCategory.join(',') : undefined

      getCommandTemplateList({
        page_index: this.currentPage,
        page_size: this.pageSize,
        search: this.searchValue,
        category
      })
        .then((res) => {
          const tableData = res?.data?.list || []
          tableData.forEach((row) => {
            row.categoryName = COMMAND_CATEGORY_NAME?.[row.category] ?? '-'
            row.createdTimeText = moment(row.created_at).format('YYYY-MM-DD HH:mm:ss')
          })
          this.tableData = tableData
          this.totalResult = res?.data?.count ?? 0
        })
        .finally(() => {
          this.loading = false
        })
    },
    onSelectChange() {
      const opsTable = this.$refs.opsTable.getVxetableRef()
      const records = [...opsTable.getCheckboxRecords(), ...opsTable.getCheckboxReserveRecords()]
      this.selectedRowKeys = records.map((i) => i.id)
    },
    onSelectRangeEnd({ records }) {
      this.selectedRowKeys = records.map((i) => i.id)
    },
    pageOrSizeChange(currentPage, pageSize) {
      this.currentPage = currentPage
      this.pageSize = pageSize
      this.updateTableData()
    },
    openModal(data) {
      this.$refs.commandTemplateModal.open(data)
    },
    deleteCommandTemplate(row) {
      this.loading = true
      deleteCommandTemplateById(row.id)
        .then((res) => {
          this.$message.success(this.$t('deleteSuccess'))
          this.updateTableData()
        })
        .finally(() => {
          this.loading = false
        })
    },
    async batchDelete() {
      this.$confirm({
        title: this.$t('warning'),
        content: this.$t('confirmDelete'),
        onOk: async () => {
          let successNum = 0
          let errorNum = 0
          this.loading = true
          this.loadTip = `${this.$t('deleting')}...`
          for (let i = 0; i < this.selectedRowKeys.length; i++) {
            await deleteCommandTemplateById(this.selectedRowKeys[i])
              .then(() => {
                successNum += 1
              })
              .catch(() => {
                errorNum += 1
              })
              .finally(() => {
                this.loadTip = this.$t('deletingTip', { total: this.selectedRowKeys.length, successNum, errorNum })
              })
          }
          this.loading = false
          this.loadTip = ''
          this.selectedRowKeys = []
          this.$refs.opsTable.getVxetableRef().clearCheckboxRow()
          this.$refs.opsTable.getVxetableRef().clearCheckboxReserve()
          this.$nextTick(() => {
            this.updateTableData()
          })
        },
      })
    },
    handleFilterChange(e) {
      switch (e.field) {
        case 'category':
          this.currentCategory = e?.values
          this.updateTableData()
          break
        default:
          break
      }
    }
  },
}
</script>

<style lang="less" scoped>
.command-template-management {
  background-color: #fff;
  height: 100%;
  border-radius: 6px;
  padding: 18px;

  &-header {
    display: flex;
    justify-content: space-between;
    margin-bottom: 16px;
  }
  &-pagination {
    text-align: right;
    margin-top: 8px;
  }
}
</style>
