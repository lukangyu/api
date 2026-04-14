import { Button, Card, Form, Input, InputNumber, Modal, Select, Space, Table, Tag, message } from 'antd'
import { useEffect, useState } from 'react'
import client from '../api/client'

type User = { id: number; username: string; display_name: string }
type ApiKey = {
  id: number
  key_prefix: string
  name: string
  request_limit: number
  request_count: number
  is_active: boolean
  user?: User
}

export default function ApiKeysPage() {
  const [items, setItems] = useState<ApiKey[]>([])
  const [users, setUsers] = useState<User[]>([])
  const [open, setOpen] = useState(false)
  const [plainKey, setPlainKey] = useState('')
  const [form] = Form.useForm()

  const load = async () => {
    const [keysResp, usersResp] = await Promise.all([client.get('/admin/api-keys'), client.get('/admin/users')])
    setItems(keysResp.data.items || [])
    setUsers(usersResp.data.items || [])
  }

  useEffect(() => {
    load().catch(() => undefined)
  }, [])

  const create = async () => {
    const values = await form.validateFields()
    const resp = await client.post('/admin/api-keys', values)
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
        <Table
          rowKey="id"
          dataSource={items}
          columns={[
            { title: '前缀', dataIndex: 'key_prefix' },
            { title: '名称', dataIndex: 'name' },
            {
              title: '所属用户',
              render: (_, row: ApiKey) => row.user?.display_name || row.user?.username || '-'
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
      </Card>

      <Modal open={open} onCancel={() => setOpen(false)} onOk={create} title="生成 API Key">
        <Form form={form} layout="vertical">
          <Form.Item name="user_id" label="所属用户" rules={[{ required: true }]}>
            <Select options={users.map((u) => ({ value: u.id, label: `${u.display_name || u.username} (${u.username})` }))} />
          </Form.Item>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="request_limit" label="请求上限(0为无限)">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="allowed_upstreams" label="允许上游ID列表(逗号分隔，可空)">
            <Input placeholder="例如 1,2,3" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal open={!!plainKey} onCancel={() => setPlainKey('')} onOk={() => setPlainKey('')} title="请保存明文 Key（仅显示一次）">
        <Input.TextArea value={plainKey} rows={3} readOnly />
      </Modal>
    </>
  )
}
