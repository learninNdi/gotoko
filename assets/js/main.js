
$(function () {
    /*=======================
                UI Slider Range JS
    =========================*/
    $("#slider-range").slider({
        range: true,
        min: 0,
        max: 2500,
        values: [10, 2500],
        slide: function (event, ui) {
            $("#amount").val("$" + ui.values[0] + " - $" + ui.values[1]);
        }
    });
    $("#amount").val("$" + $("#slider-range").slider("values", 0) +
        " - $" + $("#slider-range").slider("values", 1));
    
    let domShippingCalculationMsg = $("#shipping-calculation-msg")
    let shippingMessage = $("#shipping-message")
    let shippingDetailMessage = $("#shipping-detail-message")
    
    // get province
    $(".province_id").change(function() {
        provinceID = $(".province_id").val()

        $(".city_id").find("option")
            .remove()
            .end()
            .append(`<option value="">-- Pilih Kabupaten/Kota --</option>`)

        $.ajax({
            url: "/carts/cities?province_id=" + provinceID,
            method: "GET",
            success: function(result) {
                $.each(result.data, function(_, city) {
                    $(".city_id").append(`<option value="${city.city_id}">${city.city_name}</option>`)
                })
            }
        })
    });

    function calculateShippingFee() {
        let cityID = $(".city_id").val()
        let courier = $(".courier").val()
        
        $(".shipping_fee_options").find("option")
            .remove()
            .end()
            .append(`<option value="">-- Pilih Paket --</option>`)
            
        $.ajax({
            url: "/carts/calculate-shipping",
            method: "POST",
            data: {
                city_id: cityID,
                courier: courier,
            },
            success: function(result) {
                domShippingCalculationMsg.html("");
                $.each(result.data, function(_, shipping_fee_option) {
                    $(".shipping_fee_options").append(`<option value="${shipping_fee_option.service}">${shipping_fee_option.fee} (${shipping_fee_option.service})</option>`)
                })
            },
            error: function(e) {
                domShippingCalculationMsg.html(`<div class="alert alert-warning">Perhitungan ongkos kirim gagal!</div>`)
            }
        })
    }

    $(".courier").change(function(){
        let cityID = $(".city_id").val()

        if(cityID != "") calculateShippingFee()
    });

    $(".city_id").change(function(){
        let courier = $(".courier").val()

        if(courier != "") calculateShippingFee()
    });

    $(".shipping_fee_options").change(function() {
        let cityID = $(".city_id").val()
        let courier = $(".courier").val()
        let shippingFee = $(this).val()

        $.ajax({
            url: "/carts/apply-shipping",
            method: "POST",
            data: {
                shipping_package: shippingFee.split("-")[0],
                city_id: cityID,
                courier: courier
            },
            success: function(result) {
                $("#grand-total").text(result.data.grand_total)
            },
            error: function(e) {
                domShippingCalculationMsg.html(`<div class="alert alert-warning">Pemilihan paket ongkir gagal!</div>`)
            }
        });
    });

    $("#checkout-button").click(function(e) {
        shippingMessage.html("")
        shippingDetailMessage.html("")
        
        let provinceID = $(".province_id").val()
        let cityID = $(".city_id").val()
        let courier = $(".courier").val()
        let shippingFee = $(".shipping_fee_options").val()

        let firstName = $("#first_name").val()
        let lastName = $("#last_name").val()
        let address1 = $("#address1").val()
        let postCode = $("#post_code").val()
        let phone = $("#phone").val()
        let email = $("#email").val()

        if(provinceID == "" || cityID == "" || courier == "" || shippingFee == "" ||
            firstName == "" || lastName == "" || address1 == "" || postCode == ""
            || phone == "" || email == ""
        ) {
            e.preventDefault();            

            if(courier == "") shippingMessage.append(`<div class="alert alert-warning">Kurir belum dipilih</div>`)
            if(provinceID == "") shippingMessage.append(`<div class="alert alert-warning">Provinsi belum dipilih</div>`)
            if(cityID == "") shippingMessage.append(`<div class="alert alert-warning">Kabupaten/Kota belum dipilih</div>`)
            if(shippingFee == "") shippingMessage.append(`<div class="alert alert-warning">Paket belum dipilih</div>`)
            if(firstName == "") shippingDetailMessage.append(`<div class="alert alert-warning">Nama depan belum dipilih</div>`)
            if(lastName == "") shippingDetailMessage.append(`<div class="alert alert-warning">Nama belakang belum dipilih</div>`)
            if(address1 == "") shippingDetailMessage.append(`<div class="alert alert-warning">Alamat 1 depan belum dipilih</div>`)
            if(postCode == "") shippingDetailMessage.append(`<div class="alert alert-warning">Kode pos depan belum dipilih</div>`)
            if(phone == "") shippingDetailMessage.append(`<div class="alert alert-warning">Telepon depan belum dipilih</div>`)
            if(email == "") shippingDetailMessage.append(`<div class="alert alert-warning">Email depan belum dipilih</div>`)
        }

        // $.ajax({
        //     url: "/orders/checkout",
        //     method: "POST",
        //     success: function (result) {
        //         console.log("ada")
        //     },
        // });
    });
});
