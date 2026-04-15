import { Alert, Button, Card, DatePicker, Empty, Input, Space, Table, Tag } from 'antd'
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
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
  api_key_name: string
  user_name: string
  upstream_name: string
  user_id: number
  upstream_id: number
  created_at: string
}

export default function LogsPage() {
  const navigate = useNavigate()
  const [items, setItems] = useState<LogRow[]>([])
  const [status, setStatus] = useState('')
  const [range, setRange] = useState<[Dayjs, Dayjs] | null>(null)
  const [loading, setLoading] = useState(false)

  const load = async () => {
    setLoading(true)
    try {
      const params: Record<string, string> = {}
      if (status) params.status_code = status
      if (range) {
        params.from = range[0].toISOString()
        params.to = range[1].toISOString()
      }
      const resp = await client.get('/admin/logs', { params })
      setItems(resp.data.items || [])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load().catch(() => undefined)
  }, [])

  const emptyState = (
    <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="还没有请求日志。先复制一个上游示例请求打通链路。">
      <Space wrap>
        <Button type="primary" onClick={() => navigate('/upstreams')}>
          去查看上游示例
        </Button>
        <Button onClick={() => navigate('/')}>回到快速开始</Button>
      </Space>
    </Empty>
  )

  return (
    <Card title="请求日志">
      <Space direction="vertical" size={16} style={{ width: '100%' }}>
        {!items.length && (
          <Alert
            type="info"
            showIcon
            message="日志为空通常表示还没发起第一笔请求"
            description="如果你已经配置了上游和员工 Key，下一步去上游页面复制示例 curl 发一次请求，再回来看这里。"
          />
        )}

        <Space wrap>
          <Input placeholder="状态码过滤" value={status} onChange={(e) => setStatus(e.target.value)} style={{ width: 160 }} />
          <DatePicker.RangePicker
            showTime
            value={range as any}
            onChange={(v) => setRange((v as any) || null)}
            presets={[{ label: '最近24小时', value: [dayjs().subtract(1, 'day'), dayjs()] }]}
          />
          <Button onClick={() => load()}>查询</Button>
        </Space>

        <Table
          rowKey="id"
          loading={loading}
          dataSource={items}
          locale={{ emptyText: emptyState }}
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
            { title: '用户', dataIndex: 'user_name', render: (v: string, row: LogRow) => v || `用户#${row.user_id}` },
            { title: 'Key', dataIndex: 'api_key_name', render: (v: string) => v || '-' },
            { title: '上游', dataIndex: 'upstream_name', render: (v: string, row: LogRow) => v || `上游#${row.upstream_id}` }
          ]}
        />
      </Space>
    </Card>
  )
}
