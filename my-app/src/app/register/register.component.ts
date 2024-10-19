// src/app/register/register.component.ts
import { Component } from '@angular/core';
import { HttpClient } from '@angular/common/http';

@Component({
  selector: 'app-register',
  templateUrl: './register.component.html',
  styleUrls: ['./register.component.css']
})
export class RegisterComponent {
  username: string = '';
  password: string = '';

  constructor(private http: HttpClient) {}

  onSubmit() {
    const user = {
      username: this.username,
      password: this.password
    };

    this.http.post('http://localhost:4300/register', user).subscribe(
        response => {
          console.log('User registered successfully!', response);
        },
        error => {
          console.error('Error registering user:', error);
        }
    );
  }

}
