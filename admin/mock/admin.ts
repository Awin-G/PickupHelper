import { defineFakeRoute } from "vite-plugin-fake-server/client";

export default defineFakeRoute([
  // ====== 包裹模块 ======
  {
    url: "/api/v1/parcels/scan-in",
    method: "post",
    response: () => ({
      code: 0,
      message: "success",
      data: {
        parcel_id: 10086,
        pickup_code: "884128",
        shelf_code: "A-01",
        storage_time: "2026-07-09T17:00:00+08:00"
      }
    })
  },
  {
    url: "/api/v1/parcels",
    method: "get",
    response: () => ({
      code: 0,
      message: "success",
      data: {
        list: [
          {
            id: 1,
            tracking_no: "SF1234567890",
            courier_company: "顺丰速运",
            shelf_code: "A-01",
            pickup_code: "884128",
            receiver_phone: "138****0001",
            receiver_name: "张三",
            status: 1,
            status_text: "待取件",
            storage_time: "2026-07-08T10:00:00+08:00",
            pickup_time: null,
            notify_count: 0,
            weight: 1.5,
            is_fragile: false
          },
          {
            id: 2,
            tracking_no: "YT9876543210",
            courier_company: "圆通速递",
            shelf_code: "B-02",
            pickup_code: "559231",
            receiver_phone: "139****0002",
            receiver_name: "李四",
            status: 2,
            status_text: "已取件",
            storage_time: "2026-07-07T14:30:00+08:00",
            pickup_time: "2026-07-08T09:15:00+08:00",
            notify_count: 1,
            weight: 0.8,
            is_fragile: true
          },
          {
            id: 3,
            tracking_no: "ZT5554443330",
            courier_company: "中通快递",
            shelf_code: "C-03",
            pickup_code: "331789",
            receiver_phone: "137****0003",
            receiver_name: "王五",
            status: 3,
            status_text: "滞留",
            storage_time: "2026-07-04T08:00:00+08:00",
            pickup_time: null,
            notify_count: 3,
            weight: 2.0,
            is_fragile: false
          },
          {
            id: 4,
            tracking_no: "JD1112223330",
            courier_company: "京东物流",
            shelf_code: "A-01",
            pickup_code: "445566",
            receiver_phone: "136****0004",
            receiver_name: "赵六",
            status: 4,
            status_text: "已退件",
            storage_time: "2026-06-28T16:00:00+08:00",
            pickup_time: null,
            return_time: "2026-07-06T10:00:00+08:00",
            notify_count: 5,
            weight: 3.2,
            is_fragile: false
          },
          {
            id: 5,
            tracking_no: "EMS7778889990",
            courier_company: "EMS",
            shelf_code: "B-02",
            pickup_code: "998877",
            receiver_phone: "135****0005",
            receiver_name: "孙七",
            status: 5,
            status_text: "异常",
            storage_time: "2026-07-09T06:00:00+08:00",
            pickup_time: null,
            notify_count: 0,
            weight: 0.5,
            is_fragile: false
          }
        ],
        total: 5,
        page: 1,
        page_size: 20
      }
    })
  },
  {
    url: "/api/v1/parcels/batch-in",
    method: "post",
    response: () => ({
      code: 0,
      message: "success",
      data: { batch_id: "BATCH20260709001", total: 50, status: "pending" }
    })
  },

  // ====== 取件核销 ======
  {
    url: "/api/v1/pickup/verify",
    method: "post",
    response: () => ({
      code: 0,
      message: "success",
      data: {
        parcel_id: 10086,
        tracking_no: "SF1234567890",
        pickup_time: "2026-07-09T17:05:00+08:00",
        operator_type: 1
      }
    })
  },
  {
    url: "/api/v1/pickup/logs",
    method: "get",
    response: () => ({
      code: 0,
      message: "success",
      data: {
        list: [
          {
            id: 1,
            parcel_id: 2,
            operator_id: 1,
            operator_type: 1,
            verification_method: 1,
            created_at: "2026-07-08T09:15:00+08:00"
          },
          {
            id: 2,
            parcel_id: 6,
            operator_id: 1,
            operator_type: 1,
            verification_method: 2,
            created_at: "2026-07-07T16:30:00+08:00"
          }
        ],
        total: 2,
        page: 1,
        page_size: 20
      }
    })
  },

  // ====== 货架管理 ======
  {
    url: "/api/v1/shelves",
    method: "get",
    response: () => ({
      code: 0,
      message: "success",
      data: {
        list: [
          {
            id: 1,
            station_id: 1,
            shelf_code: "A-01",
            row_num: 4,
            col_num: 5,
            current_capacity: 15,
            max_capacity: 20,
            occupancy_rate: 0.75,
            remark: "靠门"
          },
          {
            id: 2,
            station_id: 1,
            shelf_code: "B-02",
            row_num: 4,
            col_num: 5,
            current_capacity: 18,
            max_capacity: 20,
            occupancy_rate: 0.9,
            remark: "中间"
          },
          {
            id: 3,
            station_id: 1,
            shelf_code: "C-03",
            row_num: 4,
            col_num: 5,
            current_capacity: 8,
            max_capacity: 20,
            occupancy_rate: 0.4,
            remark: "靠窗"
          }
        ],
        total: 3,
        page: 1,
        page_size: 20
      }
    })
  },
  {
    url: "/api/v1/shelves",
    method: "post",
    response: () => ({
      code: 0,
      message: "success",
      data: {
        id: 4,
        station_id: 1,
        shelf_code: "D-04",
        row_num: 4,
        col_num: 5,
        current_capacity: 0,
        max_capacity: 20,
        occupancy_rate: 0,
        remark: ""
      }
    })
  },
  {
    url: "/api/v1/shelves/occupancy",
    method: "get",
    response: () => ({
      code: 0,
      message: "success",
      data: {
        station_id: 1,
        shelves: [
          {
            shelf_code: "A-01",
            row_num: 4,
            col_num: 5,
            current_capacity: 15,
            max_capacity: 20,
            occupancy_rate: 0.75,
            heat_level: 3
          },
          {
            shelf_code: "B-02",
            row_num: 4,
            col_num: 5,
            current_capacity: 18,
            max_capacity: 20,
            occupancy_rate: 0.9,
            heat_level: 4
          },
          {
            shelf_code: "C-03",
            row_num: 4,
            col_num: 5,
            current_capacity: 8,
            max_capacity: 20,
            occupancy_rate: 0.4,
            heat_level: 2
          }
        ],
        total_used: 41,
        total_max: 60
      }
    })
  },

  // ====== 代取订单 ======
  {
    url: "/api/v1/proxy/my-orders",
    method: "get",
    response: () => ({
      code: 0,
      message: "success",
      data: {
        list: [
          {
            id: 1,
            parcel_id: 1,
            station_name: "主校区驿站",
            publisher_nickname: "张三",
            taker_nickname: "跑腿小王",
            reward_amount: 5.0,
            status: 1,
            status_text: "待接单",
            deadline: "2026-07-09T20:00:00+08:00",
            created_at: "2026-07-09T10:00:00+08:00"
          },
          {
            id: 2,
            parcel_id: 2,
            station_name: "主校区驿站",
            publisher_nickname: "李四",
            taker_nickname: "跑腿小张",
            reward_amount: 3.0,
            status: 2,
            status_text: "配送中",
            deadline: "2026-07-09T18:00:00+08:00",
            created_at: "2026-07-09T09:00:00+08:00"
          },
          {
            id: 3,
            parcel_id: 3,
            station_name: "主校区驿站",
            publisher_nickname: "王五",
            taker_nickname: "跑腿小李",
            reward_amount: 4.5,
            status: 4,
            status_text: "已完成",
            deadline: "2026-07-08T20:00:00+08:00",
            delivery_time: "2026-07-08T19:30:00+08:00",
            created_at: "2026-07-08T14:00:00+08:00"
          }
        ],
        total: 3,
        page: 1,
        page_size: 20
      }
    })
  },

  // ====== 用户管理 ======
  {
    url: "/api/v1/users",
    method: "get",
    response: () => ({
      code: 0,
      message: "success",
      data: {
        list: [
          {
            id: 1,
            phone: "138****0001",
            nickname: "张三",
            avatar: "",
            user_type: 1,
            runner_status: 0,
            credit_score: 100,
            is_blacklisted: false,
            created_at: "2026-01-15T08:00:00+08:00"
          },
          {
            id: 2,
            phone: "139****0002",
            nickname: "李四",
            avatar: "",
            user_type: 2,
            runner_status: 2,
            credit_score: 95,
            is_blacklisted: false,
            created_at: "2026-02-20T10:00:00+08:00"
          },
          {
            id: 3,
            phone: "137****0003",
            nickname: "王五",
            avatar: "",
            user_type: 1,
            runner_status: 1,
            credit_score: 80,
            is_blacklisted: true,
            created_at: "2026-03-10T14:00:00+08:00"
          },
          {
            id: 4,
            phone: "136****0004",
            nickname: "赵六",
            avatar: "",
            user_type: 2,
            runner_status: 2,
            credit_score: 70,
            is_blacklisted: false,
            created_at: "2026-04-05T09:00:00+08:00"
          }
        ],
        total: 4,
        page: 1,
        page_size: 20
      }
    })
  },
  {
    url: "/api/v1/user/runner/applications",
    method: "get",
    response: () => ({
      code: 0,
      message: "success",
      data: {
        list: [
          {
            id: 1,
            user_id: 5,
            real_name: "小明",
            phone: "135****0005",
            student_id: "2024001",
            id_card_image: "https://via.placeholder.com/300x200",
            status: 1,
            audit_remark: "",
            created_at: "2026-07-08T15:00:00+08:00",
            updated_at: "2026-07-08T15:00:00+08:00"
          },
          {
            id: 2,
            user_id: 6,
            real_name: "小红",
            phone: "134****0006",
            student_id: "2024002",
            id_card_image: "https://via.placeholder.com/300x200",
            status: 2,
            audit_remark: "审核通过",
            created_at: "2026-07-07T10:00:00+08:00",
            updated_at: "2026-07-08T09:00:00+08:00"
          },
          {
            id: 3,
            user_id: 7,
            real_name: "小刚",
            phone: "133****0007",
            student_id: "2024003",
            id_card_image: "https://via.placeholder.com/300x200",
            status: 3,
            audit_remark: "证件不清晰",
            created_at: "2026-07-06T11:00:00+08:00",
            updated_at: "2026-07-07T14:00:00+08:00"
          }
        ],
        total: 3,
        page: 1,
        page_size: 20
      }
    })
  },
  {
    url: "/api/v1/user/runner/applications/",
    method: "put",
    response: () => ({
      code: 0,
      message: "success",
      data: { id: 1, status: 2, audit_remark: "审核通过" }
    })
  },

  // ====== 统计模块 ======
  {
    url: "/api/v1/stats/dashboard",
    method: "get",
    response: () => ({
      code: 0,
      message: "success",
      data: {
        date: "2026-07-09",
        station_id: 1,
        today_inbound: 28,
        today_outbound: 15,
        pending_count: 42,
        delayed_count: 8,
        abnormal_count: 2,
        proxy_active: 6,
        shelf_usage_rate: 0.68
      }
    })
  },
  {
    url: "/api/v1/stats/trend",
    method: "get",
    response: () => ({
      code: 0,
      message: "success",
      data: {
        granularity: "day",
        points: [
          { date: "07-03", inbound: 20, outbound: 18, delayed: 2 },
          { date: "07-04", inbound: 25, outbound: 22, delayed: 3 },
          { date: "07-05", inbound: 18, outbound: 20, delayed: 1 },
          { date: "07-06", inbound: 30, outbound: 25, delayed: 4 },
          { date: "07-07", inbound: 22, outbound: 28, delayed: 2 },
          { date: "07-08", inbound: 35, outbound: 30, delayed: 5 },
          { date: "07-09", inbound: 28, outbound: 15, delayed: 3 }
        ]
      }
    })
  },
  {
    url: "/api/v1/stats/proxy-finance",
    method: "get",
    response: () => ({
      code: 0,
      message: "success",
      data: {
        total_orders: 45,
        completed_orders: 38,
        total_reward: 186.5,
        avg_reward: 4.91,
        by_taker: [
          {
            taker_id: 2,
            taker_nickname: "跑腿小王",
            order_count: 15,
            total_reward: 72.0
          },
          {
            taker_id: 4,
            taker_nickname: "跑腿小张",
            order_count: 12,
            total_reward: 58.5
          },
          {
            taker_id: 6,
            taker_nickname: "跑腿小李",
            order_count: 11,
            total_reward: 56.0
          }
        ]
      }
    })
  },
  {
    url: "/api/v1/stats/courier-check",
    method: "get",
    response: () => ({
      code: 0,
      message: "success",
      data: [
        {
          courier_company: "顺丰速运",
          inbound_count: 85,
          outbound_count: 72,
          delayed_count: 8,
          returned_count: 3,
          avg_storage_hours: 28.5
        },
        {
          courier_company: "圆通速递",
          inbound_count: 120,
          outbound_count: 105,
          delayed_count: 10,
          returned_count: 5,
          avg_storage_hours: 32.1
        },
        {
          courier_company: "中通快递",
          inbound_count: 95,
          outbound_count: 88,
          delayed_count: 6,
          returned_count: 2,
          avg_storage_hours: 25.8
        },
        {
          courier_company: "京东物流",
          inbound_count: 60,
          outbound_count: 55,
          delayed_count: 3,
          returned_count: 1,
          avg_storage_hours: 22.3
        },
        {
          courier_company: "EMS",
          inbound_count: 40,
          outbound_count: 35,
          delayed_count: 4,
          returned_count: 1,
          avg_storage_hours: 30.2
        }
      ]
    })
  },

  // ====== 驿站管理 ======
  {
    url: "/api/v1/stations",
    method: "get",
    response: () => ({
      code: 0,
      message: "success",
      data: {
        list: [
          {
            id: 1,
            name: "主校区驿站",
            address: "大学城中心区12栋",
            latitude: 23.1234567,
            longitude: 113.7654321,
            business_hours: "09:00-20:00",
            status: 1,
            status_text: "营业中",
            created_at: "2026-01-01T00:00:00+08:00"
          },
          {
            id: 2,
            name: "西区驿站",
            address: "大学城西区3栋",
            latitude: 23.1123456,
            longitude: 113.754321,
            business_hours: "10:00-19:00",
            status: 1,
            status_text: "营业中",
            created_at: "2026-02-15T00:00:00+08:00"
          },
          {
            id: 3,
            name: "东区驿站",
            address: "大学城东区8栋",
            latitude: 23.1345678,
            longitude: 113.7765432,
            business_hours: "09:00-21:00",
            status: 0,
            status_text: "休息中",
            created_at: "2026-03-01T00:00:00+08:00"
          }
        ],
        total: 3,
        page: 1,
        page_size: 20
      }
    })
  },
  {
    url: "/api/v1/stations",
    method: "post",
    response: () => ({
      code: 0,
      message: "success",
      data: {
        id: 4,
        name: "新驿站",
        address: "测试地址",
        latitude: 23.1,
        longitude: 113.7,
        business_hours: "09:00-20:00",
        status: 1
      }
    })
  }
]);
