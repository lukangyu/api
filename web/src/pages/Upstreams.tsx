import { Button, Card, Form, Input, InputNumber, Modal, Space, Switch, Table, message } from 'antd'
import { useEffect, useState } from 'react'
import client from '../api/client'

type Upstream = {
  id: number
  name: string
  display_name: string
  base_url: string
  auth_type: string
  auth_key: string
  auth_value: string
  timeout_seconds: number
  strip_prefix: boolean
  extra_headers: string
  is_active: boolean
  description: string
}

export default function UpstreamsPage() {
  const [items, setItems] = useState<Upstream[]>([])
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<Upstream | null>(null)
  const [form] = Form.useForm()

  const load = async () => {
    const resp = await client.get('/admin/upstreams')
    setItems(resp.data.items || [])
  }

  useEffect(() => {
    load().catch(() => undefined)
  }, [])

  const submit = async () => {
    const values = await form.validateFields()
    if (editing) {
      await client.put(`/admin/upstreams/${editing.id}`, values)
      message.success('更新成功')
    } else {
      await client.post('/admin/upstreams', values)
      message.success('创建成功')
    }
    setOpen(false)
    setEditing(null)
    form.resetFields()
    load()
  }

  const del = async (id: number) => {
    await client.delete(`/admin/upstreams/${id}`)
    message.success('删除成功')
    load()
  }

  return (
    <Card
      title="上游 API 管理"
      extra={
        <Button
          type="primary"
          onClick={() => {
            setEditing(null)
            form.resetFields()
            form.setFieldsValue({ auth_type: 'none', strip_prefix: true, timeout_seconds: 120, extra_headers: '{}' })
            setOpen(true)
          }}
        >
          新建上游
        </Button>
      }
    >
      <Table
        rowKey="id"
        dataSource={items}
        columns={[
          { title: '名称', dataIndex: 'name' },
          { title: '显示名', dataIndex: 'display_name' },
          { title: 'Base URL', dataIndex: 'base_url' },
          { title: '认证方式', dataIndex: 'auth_type' },
          {
            title: '启用',
            dataIndex: 'is_active',
            render: (v: boolean) => <Switch checked={v} disabled />
          },
          {
            title: '操作',
            render: (_, row: Upstream) => (
              <Space>
                <Button
                  size="small"
                  onClick={() => {
                    setEditing(row)
                    form.setFieldsValue(row)
                    setOpen(true)
                  }}
                >
                  编辑
                </Button>
                <Button danger size="small" onClick={() => del(row.id)}>
                  删除
                </Button>
              </Space>
            )
          }
        ]}
      />

      <Modal open={open} onCancel={() => setOpen(false)} onOk={submit} title={editing ? '编辑上游' : '新建上游'}>
        <Form form={form} layout="vertical">
          <Form.Item label="name" name="name" rules={[{ required: true }]}>
            <Input placeholder="openai / youtube / google" />
          </Form.Item>
          <Form.Item label="display_name" name="display_name" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item label="base_url" name="base_url" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item label="auth_type" name="auth_type" rules={[{ required: true }]}>
            <Input placeholder="none / bearer / header / query" />
          </Form.Item>
          <Form.Item label="auth_key" name="auth_key">
            <Input />
          </Form.Item>
          <Form.Item label="auth_value" name="auth_value">
            <Input.Password />
          </Form.Item>
          <Form.Item label="timeout_seconds" name="timeout_seconds">
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label="strip_prefix" name="strip_prefix" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item label="extra_headers(JSON)" name="extra_headers">
            <Input.TextArea rows={3} />
          </Form.Item>
          <Form.Item label="description" name="description">
            <Input.TextArea rows={2} />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  )
}
