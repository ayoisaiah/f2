**Đọc bằng ngôn ngữ khác:** [English](/README.md) | [Deutsch](/docs/README.de.md) | [Español](/docs/README.es.md) | [Français](/docs/README.fr.md) | [Português](/docs/README.pt.md) | [Русский](/docs/README.ru.md) | [繁體中文](/docs/README.zh.md) | Tiếng Việt

[![f2](https://ik.imagekit.io/turnupdev/f2_logo_02eDMiVt7.png)](https://ik.imagekit.io/turnupdev/f2_logo_02eDMiVt7.png)

[![](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat)](http://makeapullrequest.com) [![Github Actions](https://github.com/ayoisaiah/F2/actions/workflows/test.yml/badge.svg)](https://github.com/ayoisaiah/F2/actions) [![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://golang.org) [![GoReportCard](https://goreportcard.com/badge/github.com/ayoisaiah/f2)](https://goreportcard.com/report/github.com/ayoisaiah/f2) [![Go.mod version](https://img.shields.io/github/go-mod/go-version/ayoisaiah/f2.svg)](https://github.com/ayoisaiah/f2) [![LICENCE](https://img.shields.io/github/license/ayoisaiah/f2.svg)](https://github.com/ayoisaiah/f2/blob/master/LICENCE) [![Latest release](https://img.shields.io/github/release/ayoisaiah/f2.svg)](https://github.com/ayoisaiah/f2/releases/)

# F2 - Đổi tên hàng loạt trên dòng lệnh

**F2** là công cụ dòng lệnh đa nền tảng để đổi tên hàng loạt tệp và thư mục
một cách **nhanh chóng** và **an toàn**. Được viết bằng Go!

## F2 khác biệt ở điểm nào?

So với các công cụ đổi tên khác, F2 có một số ưu điểm nổi bật:

- **Mặc định chạy thử (dry run)**: F2 mặc định chỉ chạy thử để bạn xem lại các
thay đổi tên trước khi thực hiện thật.

- **Hỗ trợ biến**: F2 cho phép dùng các thuộc tính của tệp, chẳng hạn dữ liệu
EXIF của ảnh hoặc thẻ ID3 của tệp âm thanh, để bạn đổi tên linh hoạt tối đa.

- **Tùy chọn toàn diện**: Từ thay thế chuỗi đơn giản đến biểu thức chính quy
phức tạp, F2 cung cấp đầy đủ các khả năng đổi tên.

- **An toàn là trên hết**: F2 ưu tiên độ chính xác bằng cách bảo đảm mọi thao
tác đổi tên đều không xung đột và không lỗi thông qua các bước kiểm tra nghiêm ngặt.

- **Xử lý xung đột**: Mỗi thao tác đổi tên đều được kiểm tra trước khi thực thi,
và các xung đột phát hiện được có thể tự động giải quyết.

- **Hiệu năng cao**: F2 cực kỳ nhanh và hiệu quả, ngay cả khi đổi tên hàng nghìn
tệp cùng lúc.

- **Hoàn tác**: Mọi thao tác đổi tên đều có thể dễ dàng hoàn tác, giúp sửa lỗi
một cách đơn giản.

- **Tài liệu phong phú**: F2 được viết tài liệu kỹ lưỡng với các ví dụ thực tế,
rõ ràng, giúp bạn khai thác tối đa tính năng mà không bị rối.

## ⚡ Cài đặt

Nếu bạn là lập trình viên Go, có thể cài F2 bằng `go install` (yêu cầu phiên bản
1.23 trở lên):

```
go install github.com/ayoisaiah/f2/v2/cmd/f2@latest
```

Các cách cài đặt khác được [mô tả tại đây](https://f2.freshman.tech/guide/getting-started.html), hoặc bạn có thể xem
[trang phát hành](https://github.com/ayoisaiah/f2/releases) để tải về bản dựng
sẵn cho hệ điều hành của mình.

## 📃 Liên kết nhanh

- [Cài đặt](https://f2.freshman.tech/guide/getting-started.html)
- [Hướng dẫn bắt đầu](https://f2.freshman.tech/guide/tutorial.html)
- [Ví dụ thực tế](https://f2.freshman.tech/guide/organizing-image-library.html)
- [Các biến có sẵn](https://f2.freshman.tech/guide/how-variables-work.html)
- [Đổi tên theo cặp tệp](https://f2.freshman.tech/guide/pair-renaming.html)
- [Đổi tên bằng tệp CSV](https://f2.freshman.tech/guide/csv-renaming.html)
- [Sắp xếp](https://f2.freshman.tech/guide/sorting.html)
- [Giải quyết xung đột](https://f2.freshman.tech/guide/conflict-detection.html)
- [Hoàn tác lỗi đổi tên](https://f2.freshman.tech/guide/undoing-mistakes.html)
- [CHANGELOG](https://f2.freshman.tech/reference/changelog.html)

## 💻 Ảnh chụp màn hình

[![F2 có thể dùng thuộc tính Exif để sắp xếp các tệp ảnh](https://f2.freshman.tech/assets/2.D-uxLR9T.png)](https://f2.freshman.tech/assets/2.D-uxLR9T.png)

## 🤝 Đóng góp

Rất hoan nghênh báo lỗi và đề xuất tính năng! Vui lòng mở một issue trước khi
tạo pull request.

## ⚖ Giấy phép

Được tạo bởi Ayooluwa Isaiah và phát hành theo các điều khoản của [Giấy phép MIT](https://github.com/ayoisaiah/f2/blob/master/LICENCE).
