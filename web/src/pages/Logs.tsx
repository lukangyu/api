import { Card, DatePicker, Input, Space, Table, Tag } from 'antd'
import { useEffect, useState } from 'react'
import dayjs, { Dayjs } from 'dayjs'
import client from '../api/client'

type LogRow = {
  id: number
  method: string
  path: string
  status_code: number
  latency_ms: number
  request_bytes: number
  response_bytes: number
  user_id: number
  upstream_id: number
  created_at: string
}

export default function LogsPage() {
  const [items, setItems] = useState<LogRow[]>([])
  const [status, setStatus] = useState('')
  const [range, setRange] = useState<[Dayjs, Dayjs] | null>(null)

  const load = async () => {
    const params: Record<string, string> = {}
    if (status) params.status_code = status
    if (range) {
      params.from = range[0].toISOString()
      params.to = range[1].toISOString()
    }
    const resp = await client.get('/admin/logs', { params })
    setItems(resp.data.items || [])
  }

  useEffect(() => {
    load().catch(() => undefined)
  }, [])

  return (
    <Card title="请求日志">
      <Space style={{ marginBottom: 16 }}>
        <Input placeholder="状态码过滤" value={status} onChange={(e) => setStatus(e.target.value)} style={{ width: 160 }} />
        <DatePicker.RangePicker
          showTime
          value={range as any}
          onChange={(v) => setRange((v as any) || null)}
          presets={[{ label: '最近24小时', value: [dayjs().subtract(1, 'day'), dayjs()] }]}
        />
        <a onClick={() => load()}>查询</a>
      </Space>
      <Table
        rowKey="id"
        dataSource={items}
        columns={[
          { title: '时间', dataIndex: 'created_at' },
          { title: '方法', dataIndex: 'method' },
          { title: '路径', dataIndex: 'path' },
          {
            title: '状态码',
            dataIndex: 'status_code',
            render: (v: number) => <Tag color={v >= 400 ? 'red' : 'green'}>{v}</Tag>
          },
          { title: '延迟(ms)', dataIndex: 'latency_ms' },
          { title: '请求字节', dataIndex: 'request_bytes' },
          { title: '响应字节', dataIndex: 'response_bytes' },
          { title: '用户ID', dataIndex: 'user_id' },
          { title: '上游ID', dataIndex: 'upstream_id' }
        ]}
      />
    </Card>
  )
}
