import { Alert, Button, Card, Empty, Form, Input, InputNumber, Modal, Select, Space, Table, Tag, Typography, message } from 'antd'
import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import client from '../api/client'

type User = { id: number; username: string; display_name: string }
type Upstream = { id: number; name: string; display_name: string }
type ApiKey = {
  id: number
  key_prefix: string
  name: string
  request_limit: number
  request_count: number
  is_active: boolean
  allowed_upstreams?: string
  allowed_upstream_ids?: number[]
  user?: User
}

export default function ApiKeysPage() {
  const navigate = useNavigate()
  const [items, setItems] = useState<ApiKey[]>([])
  const [users, setUsers] = useState<User[]>([])
  const [upstreams, setUpstreams] = useState<Upstream[]>([])
  const [open, setOpen] = useState(false)
  const [plainKey, setPlainKey] = useState('')
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()

  const load = async () => {
    setLoading(true)
    try {
      const [keysResp, usersResp, upstreamsResp] = await Promise.all([
        client.get('/admin/api-keys'),
        client.get('/admin/users'),
        client.get('/admin/upstreams')
      ])
      setItems(keysResp.data.items || [])
      setUsers(usersResp.data.items || [])
      setUpstreams(upstreamsResp.data.items || [])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load().catch(() => undefined)
  }, [])

  const create = async () => {
    const values = await form.validateFields()
    const resp = await client.post('/admin/api-keys', {
      ...values,
      allowed_upstream_ids: values.allowed_upstream_ids || []
    })
    setPlainKey(resp.data.plain_key)
    setOpen(false)
    form.resetFields()
    message.success('创建成功，明文 Key 仅显示一次')
    load()
  }

  const revoke = async (id: number) => {
    await client.delete(`/admin/api-keys/${id}`)
    message.success('已撤销')
    load()
  }

  const upstreamNameByID = useMemo(
    () => new Map(upstreams.map((item) => [item.id, item.display_name || item.name])),
    [upstreams]
  )

  const renderAllowedUpstreams = (row: ApiKey) => {
    const ids = row.allowed_upstream_ids || []
    if (!ids.length) {
      return <Tag>全部上游</Tag>
    }
    return (
      <Space size={[4, 4]} wrap>
        {ids.map((id) => (
          <Tag key={id}>{upstreamNameByID.get(id) || `上游#${id}`}</Tag>
        ))}
      </Space>
    )
  }

  const emptyState = (
    <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="还没有 API Key。先给调用方生成一个员工 Key。">
      <Space wrap>
        <Button type="primary" onClick={() => setOpen(true)} disabled={!users.length}>
          生成第一个 Key
        </Button>
        <Button onClick={() => navigate('/upstreams')}>去配置上游</Button>
      </Space>
    </Empty>
  )

  return (
    <>
      <Card
        title="API Key 管理"
        extra={
          <Button type="primary" onClick={() => setOpen(true)}>
            生成 Key
          </Button>
        }
      >
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
          {!upstreams.length && (
            <Alert
              type="warning"
              showIcon
              message="还没有可代理的上游"
              description="你现在可以先生成员工 Key，但它暂时没有实际代理目标。更合理的顺序是先去“上游 API 管理”创建第一个上游。"
              action={
                <Button size="small" onClick={() => navigate('/upstreams')}>
                  去配置上游
                </Button>
              }
            />
          )}

          <Table
            rowKey="id"
            loading={loading}
            dataSource={items}
            locale={{ emptyText: emptyState }}
            columns={[
              { title: '前缀', dataIndex: 'key_prefix' },
              { title: '名称', dataIndex: 'name' },
              {
                title: '所属用户',
                render: (_, row: ApiKey) => row.user?.display_name || row.user?.username || '-'
              },
              {
                title: '允许访问上游',
                render: (_, row: ApiKey) => renderAllowedUpstreams(row)
              },
              {
                title: '额度',
                render: (_, row: ApiKey) => `${row.request_count}/${row.request_limit === 0 ? '∞' : row.request_limit}`
              },
              {
                title: '状态',
                dataIndex: 'is_active',
                render: (v: boolean) => <Tag color={v ? 'green' : 'red'}>{v ? '启用' : '停用'}</Tag>
              },
              {
                title: '操作',
                render: (_, row: ApiKey) => (
                  <Space>
                    <Button danger size="small" onClick={() => revoke(row.id)}>
                      撤销
                    </Button>
                  </Space>
                )
              }
            ]}
          />
        </Space>
      </Card>

      <Modal open={open} onCancel={() => setOpen(false)} onOk={create} title="生成 API Key">
        {!upstreams.length && (
          <Alert
            type="info"
            showIcon
            style={{ marginBottom: 16 }}
            message="当前没有上游"
            description="这个 Key 生成后可以先保留，等你创建好上游后再开始实际代理调用。"
          />
        )}
        <Form form={form} layout="vertical">
          <Form.Item name="user_id" label="所属用户" rules={[{ required: true }]}>
            <Select options={users.map((u) => ({ value: u.id, label: `${u.display_name || u.username} (${u.username})` }))} />
          </Form.Item>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input placeholder="例如 前端联调 / 搜索服务 / 向量服务" />
          </Form.Item>
          <Form.Item name="request_limit" label="请求上限(0为无限)">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="allowed_upstream_ids" label="允许访问的上游（可空）">
            <Select
              mode="multiple"
              allowClear
              placeholder="留空表示允许访问全部上游"
              options={upstreams.map((item) => ({
                value: item.id,
                label: `${item.display_name || item.name} (${item.name})`
              }))}
              optionFilterProp="label"
            />
          </Form.Item>
        </Form>
      </Modal>

      <Modal open={!!plainKey} onCancel={() => setPlainKey('')} onOk={() => setPlainKey('')} title="请保存明文 Key（仅显示一次）">
        <Typography.Paragraph
          copyable={{ text: plainKey, tooltips: ['复制', '已复制'] }}
          style={{ background: '#f5f5f5', padding: 12, borderRadius: 6, fontFamily: 'monospace', wordBreak: 'break-all' }}
        >
          {plainKey}
        </Typography.Paragraph>
      </Modal>
    </>
  )
}
