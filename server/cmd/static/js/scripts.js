 function submitFormFromFooter(button) {
    // Находим ближайшую форму к кнопке, на которую нажали
    var form = button.closest('.modal').querySelector('form');
    if (form) {
    // Находим кнопку внутри формы и вызываем на ней событие click
    var submitButton = form.querySelector('button[type="submit"]');
    if (submitButton) {
    submitButton.click();
} else {
    console.error('Submit button not found in form:', form);
}
} else {
    console.error('Form not found for button:', button);
}
}

    function submitFormFromSearch(button) {
    // Находим ближайшую форму к кнопке, на которую нажали
    var form = button.closest('.search-filter').querySelector('form');
    if (form) {
    // Находим кнопку внутри формы и вызываем на ней событие click
    var submitButton = form.querySelector('button[type="submit"]');
    if (submitButton) {
    submitButton.click();
} else {
    console.error('Submit button not found in form:', form);
}
} else {
    console.error('Form not found for button:', button);
}
}

 // Функция для удаления фотографии
 function removePhoto(photoId) {
     var photoElement = document.getElementById(photoId);
     var inputElement = document.getElementById("input_" + photoId);

     // Удаление элемента фотографии и скрытого поля
     photoElement.remove();
     inputElement.remove();
 }

 (() => {
     'use strict';

     // Fetch all the forms we want to apply custom Bootstrap validation styles to
     const forms = document.querySelectorAll('.needs-validation');

     // Loop over them and prevent submission
     Array.prototype.slice.call(forms).forEach((form) => {
         form.addEventListener('submit', (event) => {
             if (!form.checkValidity()) {
                 event.preventDefault();
                 event.stopPropagation();
             }
             form.classList.add('was-validated');
         }, false);
     });
 })();