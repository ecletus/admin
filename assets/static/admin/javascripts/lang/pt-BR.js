QOR.messages = jQuery.extend(true, QOR.messages, {
    action : {
        bulk: {
            pleaseSelectAnItem: "Você precisa selecionar pelo menos um item."
        },
        form: {
            areYouSure: "Esta ação pode não ser DESFEITA. Você tem certeza que deseja executar mesmo assim?"
        }
    },
    common: {
        ajaxError: 'Houve um erro de servidor. Por favor, tente novamente',
        recordNotFoundError: 'Registro não encontrado',
        confirm: {
            ok: "OK",
            cancel: "Cancelar"
        },
    },
    slideout: {
        confirm: "Você tem alterações não salvas neste slide. Se você fechar este slide, perderá todas as alterações " +
            "não salvas. Tem certeza de que deseja fechar o slide?"
    },
    datepicker: {
        // The date string format
        //format: 'DD/MM/YYYY',

        // The start view when initialized
        startView: 0, // 0 for days, 1 for months, 2 for years

        // The start day of the week
        weekStart: 1, // 0 for Sunday, 1 for Monday, 2 for Tuesday, 3 for Wednesday, 4 for Thursday, 5 for Friday, 6 for Saturday

        // Days' name of the week.
        days: 'Domingo Segunda Terça Quarta Quinta Sexta Sábado'.split(" "),

        // Shorter days' name
        daysShort: 'Dom Seg Ter Qua Qui Sex Sáb'.split(" "),

        // Shortest days' name
        daysMin: 'D S T Q Q S S'.split(" "),

        // Months' name
        months: "Janeiro Fevereiro Março Abril Maio Junho Julho Agosto Setembro Outubro Novembro Dezembro".split(" "),

        // Shorter months' name
        monthsShort: "Jan Fev Mar Abr Mai Jun Jul Ago Set Out Nov Dez".split(" "),

        datepicker: {
            title: 'Definir data',
            ok: 'OK',
            cancel: 'Cancelar'
        }
    },
    replicator: {
        undoDelete: 'Desfazer exclusão'
    }
});